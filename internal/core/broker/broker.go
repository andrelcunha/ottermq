package broker

import (
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
	Connections  map[net.Conn]*amqp.ConnectionInfo `json:"-"`
	mu           sync.Mutex                        `json:"-"`
	framer       amqp.Framer
	ManagerApi   ManagerApi
	ShuttingDown atomic.Bool
	ActiveConns  sync.WaitGroup
}

func NewBroker(config *config.Config) *Broker {
	b := &Broker{
		VHosts:      make(map[string]*vhost.VHost),
		Connections: make(map[net.Conn]*amqp.ConnectionInfo),
		config:      config,
	}
	b.VHosts["/"] = vhost.NewVhost("/")
	b.framer = &amqp.DefaultFramer{}
	b.ManagerApi = &DefaultManagerApi{b}
	return b
}

func (b *Broker) Start() {
	log.Println("OtterMQ version ", b.config.Version)
	log.Println("Broker is starting...")

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

	addr := fmt.Sprintf("%s:%s", b.config.Host, b.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start vhost: %v", err)
	}
	defer listener.Close()
	log.Printf("Started TCP listener on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("[DEBUG] Failed to accept connection:", err)
			continue
		}
		if ok := b.ShuttingDown.Load(); ok {
			log.Println("[DEBUG] Broker is shutting down, ignoring new connection")
			continue
		}
		log.Println("[DEBUG] New client waiting for connection: ", conn.RemoteAddr())
		go b.handleConnection((&configurations), conn)
	}
}

func (b *Broker) processRequest(conn net.Conn, newState *amqp.ChannelState) (any, error) {
	request := newState.MethodFrame
	channel := request.Channel
	connInfo := b.Connections[conn]
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
		switch request.MethodID {
		case uint16(amqp.CONNECTION_CLOSE):
			return b.closeConnectionRequested(conn, channel)
		case uint16(amqp.CONNECTION_CLOSE_OK):
			b.connectionCloseOk(conn)
			return nil, nil
		default:
			log.Printf("[DEBUG] Unknown connection method: %d", request.MethodID)
			return nil, fmt.Errorf("unknown connection method: %d", request.MethodID)
		}
	case uint16(amqp.CHANNEL):
		switch request.MethodID {
		case uint16(amqp.CHANNEL_OPEN):
			return b.openChannel(request, conn, channel)
		case uint16(amqp.CHANNEL_CLOSE):
			return b.closeChannel(conn, channel)
		default:
			log.Printf("[DEBUG] Unknown channel method: %d", request.MethodID)
			return nil, fmt.Errorf("unknown channel method: %d", request.MethodID)
		}

	case uint16(amqp.EXCHANGE):
		switch request.MethodID {
		case uint16(amqp.EXCHANGE_DECLARE):
			fmt.Printf("[DEBUG] Received exchange declare request: %+v\n", request)
			fmt.Printf("[DEBUG] Channel: %d\n", channel)
			content, ok := request.Content.(*amqp.ExchangeDeclareMessage)
			if !ok {
				fmt.Printf("Invalid content type for ExchangeDeclareMessage")
				return nil, fmt.Errorf("invalid content type for ExchangeDeclareMessage")
			}
			fmt.Printf("[DEBUG] Content: %+v\n", content)
			typ := content.ExchangeType
			exchangeName := content.ExchangeName

			err := vh.CreateExchange(exchangeName, vhost.ExchangeType(typ))
			if err != nil {
				return nil, err
			}
			frame := b.framer.CreateExchangeDeclareFrame(channel, request)

			b.framer.SendFrame(conn, frame)
			return nil, nil

		case uint16(amqp.EXCHANGE_DELETE):
			fmt.Printf("[DEBUG] Received exchange.delete request: %+v\n", request)
			fmt.Printf("[DEBUG] Channel: %d\n", channel)
			content, ok := request.Content.(*amqp.ExchangeDeleteMessage)
			if !ok {
				fmt.Printf("Invalid content type for ExchangeDeleteMessage")
				return nil, fmt.Errorf("invalid content type for ExchangeDeleteMessage")
			}
			fmt.Printf("[DEBUG] Content: %+v\n", content)
			exchangeName := content.ExchangeName

			err := vh.DeleteExchange(exchangeName)
			if err != nil {
				return nil, err
			}

			frame := amqp.ResponseMethodMessage{
				Channel:  channel,
				ClassID:  request.ClassID,
				MethodID: uint16(amqp.EXCHANGE_DELETE_OK),
				Content:  amqp.ContentList{},
			}.FormatMethodFrame()

			b.framer.SendFrame(conn, frame)
			return nil, nil

		default:
			return nil, fmt.Errorf("unsupported command")
		}

	case uint16(amqp.QUEUE):
		switch request.MethodID {
		case uint16(amqp.QUEUE_DECLARE):
			fmt.Printf("[DEBUG] Received queue declare request: %+v\n", request)
			content, ok := request.Content.(*amqp.QueueDeclareMessage)
			if !ok {
				fmt.Printf("Invalid content type for ExchangeDeclareMessage")
				return nil, fmt.Errorf("invalid content type for ExchangeDeclareMessage")
			}
			fmt.Printf("[DEBUG] Content: %+v\n", content)
			queueName := content.QueueName

			queue, err := vh.CreateQueue(queueName)
			if err != nil {
				return nil, err
			}

			err = vh.BindToDefaultExchange(queueName)
			if err != nil {
				fmt.Printf("[DEBUG] Error binding to default exchange: %v\n", err)
				return nil, err
			}
			messageCount := uint32(queue.Len())
			counsumerCount := uint32(0)

			frame := amqp.ResponseMethodMessage{
				Channel:  channel,
				ClassID:  request.ClassID,
				MethodID: uint16(amqp.QUEUE_DECLARE_OK),
				Content: amqp.ContentList{
					KeyValuePairs: []amqp.KeyValue{
						{
							Key:   amqp.STRING_SHORT,
							Value: queueName,
						},
						{
							Key:   amqp.INT_LONG,
							Value: messageCount,
						},
						{
							Key:   amqp.INT_LONG,
							Value: counsumerCount,
						},
					},
				},
			}.FormatMethodFrame()
			b.framer.SendFrame(conn, frame)
			return nil, nil

		case uint16(amqp.QUEUE_BIND):
			fmt.Printf("[DEBUG] Received queue bind request: %+v\n", request)
			content, ok := request.Content.(*amqp.QueueBindMessage)
			if !ok {
				fmt.Printf("Invalid content type for QueueBindMessage")
				return nil, fmt.Errorf("invalid content type for QueueBindMessage")
			}
			fmt.Printf("[DEBUG] Content: %+v\n", content)
			queue := content.Queue
			exchange := content.Exchange
			routingKey := content.RoutingKey

			err := vh.BindQueue(exchange, queue, routingKey)
			if err != nil {
				fmt.Printf("[DEBUG] Error binding to default exchange: %v\n", err)
				return nil, err
			}
			frame := amqp.ResponseMethodMessage{
				Channel:  channel,
				ClassID:  request.ClassID,
				MethodID: uint16(amqp.QUEUE_BIND_OK),
				Content:  amqp.ContentList{},
			}.FormatMethodFrame()
			b.framer.SendFrame(conn, frame)
			return nil, nil

		case uint16(amqp.QUEUE_DELETE):
			return nil, fmt.Errorf("not implemented")

		case uint16(amqp.QUEUE_UNBIND):
			return nil, fmt.Errorf("not implemented")

		default:
			return nil, fmt.Errorf("unsupported command")
		}
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
				fmt.Printf("[DEBUG] Current state after update method : %+v\n", b.getCurrentState(conn, channel))
				return nil, nil
			}
			// if the class and method are not the same as the current state,
			// it means that it stated the new publish request
			if currentState.HeaderFrame == nil && newState.HeaderFrame != nil {
				b.Connections[conn].Channels[channel].HeaderFrame = newState.HeaderFrame
				b.Connections[conn].Channels[channel].BodySize = newState.HeaderFrame.BodySize
				fmt.Printf("[DEBUG] Current state after update header: %+v\n", b.getCurrentState(conn, channel))
				return nil, nil
			}
			if currentState.Body == nil && newState.Body != nil {
				b.Connections[conn].Channels[channel].Body = newState.Body
			}
			fmt.Printf("[DEBUG] Current state after all: %+v\n", currentState)
			if currentState.MethodFrame.Content != nil && currentState.HeaderFrame != nil && currentState.BodySize > 0 && currentState.Body != nil {
				fmt.Printf("[DEBUG] All fields must be filled -> current state: %+v\n", currentState)
				if len(currentState.Body) != int(currentState.BodySize) {
					fmt.Printf("[DEBUG] Body size is not correct: %d != %d\n", len(currentState.Body), currentState.BodySize)
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
				fmt.Printf("[ERROR] Error getting message count: %v", err)
				return nil, err
			}
			if msgCount == 0 {
				// Send Basic.GetEmpty
				reserved1 := amqp.KeyValue{
					Key:   amqp.STRING_SHORT,
					Value: "",
				}
				frame := amqp.ResponseMethodMessage{
					Channel:  channel,
					ClassID:  request.ClassID,
					MethodID: uint16(amqp.BASIC_GET_EMPTY),
					Content:  amqp.ContentList{KeyValuePairs: []amqp.KeyValue{reserved1}},
				}.FormatMethodFrame()
				fmt.Printf("[DEBUG] Sending get-empty frame: %x\n", frame)
				b.framer.SendFrame(conn, frame)
				return nil, nil
			}

			// Send Basic.GetOk + header + body
			msg := vh.MsgCtrlr.GetMessage(queue)
			msgGetOk := &amqp.BasicGetOk{
				DeliveryTag:  1,
				Redelivered:  false,
				Exchange:     msg.Exchange,
				RoutingKey:   msg.RoutingKey,
				MessageCount: uint32(msgCount),
			}

			frame := amqp.ResponseMethodMessage{
				Channel:  channel,
				ClassID:  request.ClassID,
				MethodID: uint16(amqp.BASIC_GET_OK),
				Content:  *amqp.EncodeGetOkToContentList(msgGetOk),
			}.FormatMethodFrame()

			err = b.framer.SendFrame(conn, frame)
			log.Printf("[DEBUG] Sent message from queue %s: ID=%s", queue, msg.ID)

			if err != nil {
				fmt.Printf("[DEBUG] Error sending frame: %v\n", err)
				return nil, err
			}

			responseContent := amqp.ResponseContent{
				Channel: channel,
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
		// Handle transaction-related commands
		switch request.MethodID {
		case uint16(amqp.SELECT):
			// Handle transaction selection
		case uint16(amqp.COMMIT):
			// Handle transaction commit
		case uint16(amqp.ROLLBACK):
			// Handle transaction rollback
		default:
			return nil, fmt.Errorf("unsupported command")
		}
	default:
		return nil, fmt.Errorf("unsupported command")
	}
	return nil, nil
}

func (b *Broker) updateCurrentState(conn net.Conn, channel uint16, newState *amqp.ChannelState) {
	fmt.Println("Updating current state on channel ", channel)
	currentState := b.getCurrentState(conn, channel)
	b.mu.Lock()
	defer b.mu.Unlock()
	if newState.MethodFrame != nil {
		currentState.MethodFrame = newState.MethodFrame
	}
	if newState.HeaderFrame != nil {
		currentState.HeaderFrame = newState.HeaderFrame
	}
	if newState.Body != nil {
		currentState.Body = append(currentState.Body, newState.Body...)
	}
	if newState.BodySize != 0 {
		currentState.BodySize = newState.BodySize
	}
	b.Connections[conn].Channels[channel] = currentState
}

func (b *Broker) getCurrentState(conn net.Conn, channel uint16) *amqp.ChannelState {
	b.mu.Lock()
	defer b.mu.Unlock()
	fmt.Printf("[DEBUG] Getting current state for channel %d\n", channel)
	state, ok := b.Connections[conn].Channels[channel]
	if !ok {
		fmt.Printf("[DEBUG] No channel found\n")
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
