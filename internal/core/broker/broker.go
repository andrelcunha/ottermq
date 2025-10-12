package broker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/andrelcunha/ottermq/config"
	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
	"github.com/andrelcunha/ottermq/pkg/persistence"
	"github.com/andrelcunha/ottermq/pkg/persistence/implementations/json"
	"github.com/rs/zerolog/log"

	_ "github.com/andrelcunha/ottermq/internal/persistdb"
)

const (
	platform = "golang"
	product  = "OtterMQ"
)

type Broker struct {
	VHosts       map[string]*vhost.VHost
	config       *config.Config                    `json:"-"`
	listener     net.Listener                      `json:"-"`
	Connections  map[net.Conn]*amqp.ConnectionInfo `json:"-"`
	mu           sync.Mutex                        `json:"-"`
	framer       amqp.Framer
	ManagerApi   ManagerApi
	ShuttingDown atomic.Bool
	ActiveConns  sync.WaitGroup
	rootCtx      context.Context
	rootCancel   context.CancelFunc
	persist      persistence.Persistence
}

func NewBroker(config *config.Config, rootCtx context.Context, rootCancel context.CancelFunc) *Broker {
	// Create persistence layer based on config
	persistConfig := &persistence.Config{
		Type:    "json", // from config or env var
		DataDir: "data",
		Options: make(map[string]string),
	}
	persist, err := json.NewJsonPersistence(persistConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create JSON persistence")
	}

	b := &Broker{
		VHosts:      make(map[string]*vhost.VHost),
		Connections: make(map[net.Conn]*amqp.ConnectionInfo),
		config:      config,
		rootCtx:     rootCtx,
		rootCancel:  rootCancel,
		persist:     persist,
	}
	b.VHosts["/"] = vhost.NewVhost("/", config.QueueBufferSize, b.persist)
	b.framer = &amqp.DefaultFramer{}
	b.ManagerApi = &DefaultManagerApi{b}
	return b
}

func (b *Broker) Start() error {
	Logo()
	log.Info().Str("version", b.config.Version).Msg("OtterMQ")
	log.Info().Msg("Broker is starting...")

	configurations := b.setConfigurations()

	addr := fmt.Sprintf("%s:%s", b.config.Host, b.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start listener")
	}
	b.listener = listener
	defer listener.Close()
	log.Info().Str("addr", addr).Msg("Started TCP listener")

	return b.acceptLoop(configurations)
}

func Logo() {
	// TODO: create a better logo: 🦦
	fmt.Print(`
	
 OOOOO  tt    tt                  MM    MM  QQQQQ 
OO   OO tt    tt      eee  rr rr  MMM  MMM QQ   QQ
OO   OO tttt  tttt  ee   e rrr  r MM MM MM QQ   QQ
OO   OO tt    tt    eeeee  rr     MM    MM QQ  QQ 
 OOOO0   tttt  tttt  eeeee rr     MM    MM  QQQQ Q

`)
}
func (b *Broker) setConfigurations() map[string]any {
	capabilities := map[string]any{
		"basic.nack":             true,
		"connection.blocked":     true,
		"consumer_cancel_notify": true,
		"publisher_confirms":     true,
	}

	serverProperties := map[string]any{
		"capabilities": capabilities,
		"product":      product,
		"version":      b.config.Version,
		"platform":     platform,
	}

	configurations := map[string]any{
		"mechanisms":        []string{"PLAIN"},
		"locales":           []string{"en_US"},
		"serverProperties":  serverProperties,
		"heartbeatInterval": b.config.HeartbeatIntervalMax,
		"frameMax":          b.config.FrameMax,
		"channelMax":        b.config.ChannelMax,
		"ssl":               b.config.Ssl,
		"protocol":          "AMQP 0-9-1",
	}
	return configurations
}

func (b *Broker) acceptLoop(configurations map[string]any) error {
	for {
		conn, err := b.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || b.rootCtx.Err() != nil {
				log.Info().Err(err).Msg("Listener closed or context canceled")
				return err
			}
			log.Debug().Err(err).Msg("Accept failed")
			continue
		}
		if ok := b.ShuttingDown.Load(); ok {
			log.Debug().Msg("Broker is shutting down, ignoring new connection")
			continue
		}
		log.Debug().Str("client", conn.RemoteAddr().String()).Msg("New client waiting for connection")
		connCtx, connCancel := context.WithCancel(b.rootCtx)
		defer connCancel()
		connInfo, err := b.framer.Handshake(&configurations, conn, connCtx)
		if err != nil {
			log.Info().Err(err).Msg("Handshake failed")
			continue
		}
		b.registerConnection(conn, connInfo)
		go b.monitorConnectionLifecycle(conn, connInfo.Client)

		go b.handleConnection(conn, connInfo)
	}
}

func (b *Broker) monitorConnectionLifecycle(conn net.Conn, client *amqp.AmqpClient) {
	<-client.Ctx.Done()
	log.Info().Str("client", conn.RemoteAddr().String()).Msg("Connection closed")
	b.cleanupConnection(conn)
}

func (b *Broker) processRequest(conn net.Conn, newState *amqp.ChannelState) (any, error) {
	request := newState.MethodFrame
	connInfo, exist := b.Connections[conn]
	if !exist {
		// connection terminated while processing the request
		log.Info().Str("client", conn.RemoteAddr().String()).Msg("Connection closed")
		return nil, nil
	}
	vh := b.VHosts[connInfo.VHostName]
	if ok := b.ShuttingDown.Load(); ok {
		if request.ClassID != uint16(amqp.CONNECTION) ||
			(request.MethodID != uint16(amqp.CONNECTION_CLOSE) &&
				request.MethodID != uint16(amqp.CONNECTION_CLOSE_OK)) {
			return nil, nil
		}
	}

	switch request.ClassID {
	case uint16(amqp.CONNECTION):
		return b.connectionHandler(request, conn)
	case uint16(amqp.CHANNEL):
		return b.channelHandler(request, conn)
	case uint16(amqp.EXCHANGE):
		return b.exchangeHandler(request, vh, conn)
	case uint16(amqp.QUEUE):
		return b.queueHandler(request, vh, conn)
	case uint16(amqp.BASIC):
		return b.basicHandler(newState, vh, conn)
	case uint16(amqp.TX):
		return b.txHandler(request)
	default:
		return nil, fmt.Errorf("unsupported command")
	}
}

func (b *Broker) getCurrentState(conn net.Conn, channel uint16) *amqp.ChannelState {
	b.mu.Lock()
	defer b.mu.Unlock()
	log.Debug().Uint16("channel", channel).Msg("Getting current state for channel")
	state, ok := b.Connections[conn].Channels[channel]
	if !ok {
		log.Debug().Msg("No channel found")
		return nil
	}
	return state
}

func (b *Broker) GetVHost(vhostName string) *vhost.VHost {
	b.mu.Lock()
	defer b.mu.Unlock()
	if vhost, ok := b.VHosts[vhostName]; ok {
		return vhost
	}
	return nil
}

func (b *Broker) Shutdown() {
	for conn := range b.Connections {
		conn.Close()
	}
}
