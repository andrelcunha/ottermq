package broker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/andrelcunha/ottermq/config"
	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"

	_ "github.com/andrelcunha/ottermq/internal/core/persistdb"
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
}

func NewBroker(config *config.Config, rootCtx context.Context, rootCancel context.CancelFunc) *Broker {
	b := &Broker{
		VHosts:      make(map[string]*vhost.VHost),
		Connections: make(map[net.Conn]*amqp.ConnectionInfo),
		config:      config,
		rootCtx:     rootCtx,
		rootCancel:  rootCancel,
	}
	b.VHosts["/"] = vhost.NewVhost("/")
	b.framer = &amqp.DefaultFramer{}
	b.ManagerApi = &DefaultManagerApi{b}
	return b
}

func (b *Broker) Start() error {
	b.Logo()
	log.Printf("OtterMQ version %s", b.config.Version)
	log.Printf("Broker is starting...")

	configurations := b.setConfigurations()

	addr := fmt.Sprintf("%s:%s", b.config.Host, b.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start vhost: %v", err)
	}
	b.listener = listener
	defer listener.Close()
	log.Printf("Started TCP listener on %s", addr)

	return b.acceptLoop(configurations)
}

func (b *Broker) Logo() {
	// TODO: create a better logo: 🦦
	log.Printf(
		`

  oooooo     o8     o8                        oooo     oooo  oooooo  
o888   888o o888oo o888oo oooooooo8 oo ooooo   8888o   888 o888  888o
888     888  888    888  888ooooo8   888   888 88 888o8 88 888    888
888o   o888  888    888  888         888       88  888  88 888o  o888
   88o88     888o   888o  88ooo888  o888o     o88o  8  o88o  88oo88  
                                                                 88o8
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
				log.Printf("[INFO] Listener closed or context canceled: %v", err)
				return err
			}
			log.Printf("[DEBUG] Accept failed: %v", err)
			continue
		}
		if ok := b.ShuttingDown.Load(); ok {
			log.Println("[DEBUG] Broker is shutting down, ignoring new connection")
			continue
		}
		log.Println("[DEBUG] New client waiting for connection: ", conn.RemoteAddr())
		connCtx, connCancel := context.WithCancel(b.rootCtx)
		defer connCancel()
		connInfo, err := b.framer.Handshake(&configurations, conn, connCtx)
		if err != nil {
			log.Printf("[INFO] Handshake failed: %v", err)
			continue
		}
		b.registerConnection(conn, connInfo)
		go b.monitorConnectionLifecycle(conn, connInfo.Client)

		go b.handleConnection(conn, connInfo)
	}
}

func (b *Broker) monitorConnectionLifecycle(conn net.Conn, client *amqp.AmqpClient) {
	<-client.Ctx.Done()
	log.Printf("[INFO] Connection %s closed", conn.RemoteAddr())
	b.cleanupConnection(conn)
}

func (b *Broker) processRequest(conn net.Conn, newState *amqp.ChannelState) (any, error) {
	request := newState.MethodFrame
	connInfo, exist := b.Connections[conn]
	if !exist {
		return nil, fmt.Errorf("connection not found")
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
		switch request.MethodID {
		case uint16(amqp.BASIC_QOS):
		case uint16(amqp.BASIC_CONSUME):
		case uint16(amqp.BASIC_CANCEL):
		case uint16(amqp.BASIC_PUBLISH):
			channel := request.Channel
			currentState := b.getCurrentState(conn, channel)
			if currentState == nil {
				return nil, fmt.Errorf("channel not found")
			}
			if currentState.MethodFrame != newState.MethodFrame {
				b.Connections[conn].Channels[channel].MethodFrame = newState.MethodFrame
				log.Printf("[DEBUG] Current state after update method : %+v\n", b.getCurrentState(conn, channel))
				return nil, nil
			}
			// if the class and method are not the same as the current state,
			// it means that it stated the new publish request
			if currentState.HeaderFrame == nil && newState.HeaderFrame != nil {
				b.Connections[conn].Channels[channel].HeaderFrame = newState.HeaderFrame
				b.Connections[conn].Channels[channel].BodySize = newState.HeaderFrame.BodySize
				log.Printf("[DEBUG] Current state after update header: %+v\n", b.getCurrentState(conn, channel))
				return nil, nil
			}
			if currentState.Body == nil && newState.Body != nil {
				b.Connections[conn].Channels[channel].Body = newState.Body
			}
			log.Printf("[DEBUG] Current state after all: %+v\n", currentState)
			if currentState.MethodFrame.Content != nil && currentState.HeaderFrame != nil && currentState.BodySize > 0 && currentState.Body != nil {
				log.Printf("[DEBUG] All fields must be filled -> current state: %+v\n", currentState)
				if len(currentState.Body) != int(currentState.BodySize) {
					log.Printf("[DEBUG] Body size is not correct: %d != %d\n", len(currentState.Body), currentState.BodySize)
					return nil, fmt.Errorf("body size is not correct: %d != %d", len(currentState.Body), currentState.BodySize)
				}
				publishRequest := currentState.MethodFrame.Content.(*amqp.BasicPublishMessage)
				exchanege := publishRequest.Exchange
				routingKey := publishRequest.RoutingKey
				body := currentState.Body
				props := currentState.HeaderFrame.Properties
				_, err := vh.MsgCtrlr.Publish(exchanege, routingKey, body, props)
				if err == nil {
					log.Printf("[DEBUG] Published message to exchange=%s, routingKey=%s, body=%s", exchanege, routingKey, string(body))
					b.Connections[conn].Channels[channel] = &amqp.ChannelState{}
				}
				return nil, err

			}
		case uint16(amqp.BASIC_GET):
			getMsg := request.Content.(*amqp.BasicGetMessage)
			queue := getMsg.Queue
			msgCount, err := vh.MsgCtrlr.GetMessageCount(queue)
			if err != nil {
				log.Printf("[ERROR] Error getting message count: %v", err)
				return nil, err
			}
			if msgCount == 0 {
				frame := b.framer.CreateBasicGetEmptyFrame(request)
				b.framer.SendFrame(conn, frame)
				return nil, nil
			}

			// Send Basic.GetOk + header + body
			msg := vh.MsgCtrlr.GetMessage(queue)

			frame := b.framer.CreateBasicGetOkFrame(request, msg.Exchange, msg.RoutingKey, uint32(msgCount))
			err = b.framer.SendFrame(conn, frame)
			log.Printf("[DEBUG] Sent message from queue %s: ID=%s", queue, msg.ID)

			if err != nil {
				log.Printf("[DEBUG] Error sending frame: %v\n", err)
				return nil, err
			}

			responseContent := amqp.ResponseContent{
				Channel: request.Channel,
				ClassID: request.ClassID,
				Weight:  0,
				Message: *msg,
			}
			// Header
			frame = responseContent.FormatHeaderFrame()
			b.framer.SendFrame(conn, frame)
			// Body
			frame = responseContent.FormatBodyFrame()
			b.framer.SendFrame(conn, frame)
			return nil, nil

		case uint16(amqp.BASIC_ACK):
			return nil, fmt.Errorf("not implemented")

		case uint16(amqp.BASIC_REJECT):
		case uint16(amqp.BASIC_RECOVER_ASYNC):
		case uint16(amqp.BASIC_RECOVER):
		default:
			return nil, fmt.Errorf("unsupported command")
		}
	case uint16(amqp.TX):
		return b.txHandler(request)
	default:
		return nil, fmt.Errorf("unsupported command")
	}
	return nil, nil
}

// func (b *Broker) updateCurrentState(conn net.Conn, channel uint16, newState *amqp.ChannelState) {
// 	fmt.Println("Updating current state on channel ", channel)
// 	currentState := b.getCurrentState(conn, channel)
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
// 	if newState.MethodFrame != nil {
// 		currentState.MethodFrame = newState.MethodFrame
// 	}
// 	if newState.HeaderFrame != nil {
// 		currentState.HeaderFrame = newState.HeaderFrame
// 	}
// 	if newState.Body != nil {
// 		currentState.Body = append(currentState.Body, newState.Body...)
// 	}
// 	if newState.BodySize != 0 {
// 		currentState.BodySize = newState.BodySize
// 	}
// 	b.Connections[conn].Channels[channel] = currentState
// }

func (b *Broker) getCurrentState(conn net.Conn, channel uint16) *amqp.ChannelState {
	b.mu.Lock()
	defer b.mu.Unlock()
	log.Printf("[DEBUG] Getting current state for channel %d\n", channel)
	state, ok := b.Connections[conn].Channels[channel]
	if !ok {
		log.Printf("[DEBUG] No channel found\n")
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
