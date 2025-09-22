package broker

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/andrelcunha/ottermq/config"
	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/amqp/shared"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"

	_ "github.com/andrelcunha/ottermq/internal/core/persistdb"
)

var (
	version = "0.6.0-alpha"
)

const (
	platform = "golang"
	product  = "OtterMQ"
)

type Broker struct {
	VHosts      map[string]*vhost.VHost
	config      *config.Config               `json:"-"`
	Connections map[net.Conn]*ConnectionInfo `json:"-"`
	mu          sync.Mutex                   `json:"-"`
}

func NewBroker(config *config.Config) *Broker {
	b := &Broker{
		VHosts:      make(map[string]*vhost.VHost),
		Connections: make(map[net.Conn]*ConnectionInfo),
		config:      config,
	}
	b.VHosts["/"] = vhost.NewVhost("/")
	return b
}

func (b *Broker) Start() {
	capabilities := map[string]any{
		"basic.nack":             true,
		"connection.blocked":     true,
		"consumer_cancel_notify": true,
		"publisher_confirms":     true,
	}

	serverProperties := map[string]interface{}{
		"capabilities": capabilities,
		"product":      product,
		"version":      version,
		"platform":     platform,
	}

	configurations := map[string]interface{}{
		"mechanisms":        []string{"PLAIN"},
		"locales":           []string{"en_US"},
		"serverProperties":  serverProperties,
		"heartbeatInterval": b.config.HeartbeatIntervalMax,
		"frameMax":          b.config.FrameMax,
		"channelMax":        b.config.ChannelMax,
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
			log.Println("Failed to accept connection:", err)
			continue
		}
		log.Println("New client waiting for connection: ", conn.RemoteAddr())
		go b.handleConnection((&configurations), conn)
	}
}

func (b *Broker) ParseFrame(configurations *map[string]interface{}, conn net.Conn, currentChannel uint16, frame []byte) (interface{}, error) {
	if len(frame) < 7 {
		return nil, fmt.Errorf("frame too short")
	}

	frameType := frame[0]
	channel := binary.BigEndian.Uint16(frame[1:3])
	payloadSize := binary.BigEndian.Uint32(frame[3:7])
	if len(frame) < int(7+payloadSize) {
		return nil, fmt.Errorf("frame too short")
	}

	payload := frame[7:]

	switch frameType {
	case byte(amqp.TYPE_METHOD):
		log.Printf("[DEBUG] Received METHOD frame on channel %d\n", channel)
		request, err := shared.ParseMethodFrame(configurations, channel, payload)
		if err != nil {
			return nil, fmt.Errorf("failed to parse method frame: %v", err)
		}
		return request, nil

	case byte(amqp.TYPE_HEADER):
		fmt.Printf("Received HEADER frame on channel %d\n", channel)

		return shared.ParseHeaderFrame(channel, payloadSize, payload)

	case byte(amqp.TYPE_BODY):
		fmt.Printf("Received BODY frame on channel %d\n", channel)

		return shared.ParseBodyFrame(channel, payloadSize, payload)

	case byte(amqp.TYPE_HEARTBEAT):
		log.Printf("[DEBUG] Received HEARTBEAT frame on channel %d\n", channel)
		err := b.handleHeartbeat(conn)
		if err != nil {
			log.Printf("Error handling heartbeat: %v", err)
		}
		return nil, nil

	default:
		fmt.Printf("Received: %x\n", frame)
		return nil, fmt.Errorf("unknown frame type: %d", frameType)
	}
}

func (b *Broker) handleHeartbeat(conn net.Conn) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Connections[conn].LastHeartbeat = time.Now()
	return nil
}

func (b *Broker) sendHeartbeats(conn net.Conn) {
	b.mu.Lock()
	connectionInfo, ok := b.Connections[conn]
	if !ok {
		b.mu.Unlock()
		return
	}
	heartbeatInterval := connectionInfo.HeartbeatInterval
	done := connectionInfo.Done
	b.mu.Unlock()

	ticker := time.NewTicker(time.Duration(heartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			if _, ok := b.Connections[conn]; !ok {
				b.mu.Unlock()
				log.Println("Connection no longer exists in broker")
				return
			}
			b.mu.Unlock()

			// sendHearbeat(conn)
			heartbeatFrame := shared.CreateHeartbeatFrame()
			err := shared.SendFrame(conn, heartbeatFrame)
			if err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				return
			}
			log.Println("[DEBUG] Heartbeat sent")

		case <-done:
			log.Println("Stopping heartbeat  goroutine for closed connection")
			return
		}

	}
}

func (b *Broker) processRequest(conn net.Conn, newState *amqp.ChannelState) (any, error) {
	request := newState.MethodFrame
	switch request.ClassID {

	case uint16(amqp.CONNECTION):
		switch request.MethodID {

		case uint16(amqp.CONNECTION_CLOSE):
			b.cleanupConnection(conn)
			frame := amqp.ResponseMethodMessage{
				Channel:  request.Channel,
				ClassID:  uint16(amqp.CONNECTION),
				MethodID: uint16(amqp.CONNECTION_CLOSE_OK),
				Content:  amqp.ContentList{},
			}.FormatMethodFrame()
			shared.SendFrame(conn, frame)
			return nil, nil
		default:
			log.Printf("[DEBUG] Unknown connection method: %d", request.MethodID)
			return nil, fmt.Errorf("Unknown connection method: %d", request.MethodID)
		}

	case uint16(amqp.CHANNEL):
		switch request.MethodID {
		case uint16(amqp.CHANNEL_OPEN):
			fmt.Printf("[DEBUG] Received channel open request: %+v\n", request)
			channelId := request.Channel
			// Check if the channel is already open
			if b.checkChannel(conn, channelId) {
				fmt.Printf("[DEBUG] Channel %d already open\n", channelId)
				return nil, fmt.Errorf("Channel already open")
			}
			b.addChannel(conn, request)
			fmt.Printf("[DEBUG] New state added: %+v\n", b.Connections[conn].Channels[request.Channel])

			frame := amqp.ResponseMethodMessage{
				Channel:  channelId,
				ClassID:  request.ClassID,
				MethodID: uint16(amqp.CHANNEL_OPEN_OK),
				Content: amqp.ContentList{
					KeyValuePairs: []amqp.KeyValue{
						{
							Key:   amqp.INT_LONG,
							Value: uint32(0),
						},
					},
				},
			}.FormatMethodFrame()

			shared.SendFrame(conn, frame)
			return nil, nil

		case uint16(amqp.CHANNEL_CLOSE):
			channelId := request.Channel
			// check if channel is open
			if b.checkChannel(conn, channelId) {
				fmt.Printf("[DEBUG] Channel %d already open\n", channelId)
				return nil, fmt.Errorf("Channel already open")
			}
			b.removeChannel(conn, channelId)
			frame := amqp.ResponseMethodMessage{
				Channel:  channelId,
				ClassID:  uint16(amqp.CHANNEL),
				MethodID: uint16(amqp.CHANNEL_CLOSE_OK),
				Content:  amqp.ContentList{},
			}.FormatMethodFrame()
			shared.SendFrame(conn, frame)
			return nil, nil

		default:
			log.Printf("[DEBUG] Unknown channel method: %d", request.MethodID)
			return nil, fmt.Errorf("Unknown channel method: %d", request.MethodID)
		}

	case uint16(amqp.EXCHANGE):
		switch request.MethodID {
		case uint16(amqp.EXCHANGE_DECLARE):
			fmt.Printf("[DEBUG] Received exchange declare request: %+v\n", request)
			channelId := request.Channel
			fmt.Printf("[DEBUG] Channel: %d\n", channelId)
			content, ok := request.Content.(*amqp.ExchangeDeclareMessage)
			if !ok {
				fmt.Printf("Invalid content type for ExchangeDeclareMessage")
				return nil, fmt.Errorf("Invalid content type for ExchangeDeclareMessage")
			}
			fmt.Printf("[DEBUG] Content: %+v\n", content)
			typ := content.ExchangeType
			exchangeName := content.ExchangeName

			vh := b.VHosts["/"]

			err := vh.CreateExchange(exchangeName, vhost.ExchangeType(typ))
			if err != nil {
				return nil, err
			}
			frame := amqp.ResponseMethodMessage{
				Channel:  channelId,
				ClassID:  request.ClassID,
				MethodID: uint16(amqp.EXCHANGE_DECLARE_OK),
				Content:  amqp.ContentList{},
			}.FormatMethodFrame()

			shared.SendFrame(conn, frame)
			return nil, nil

		case uint16(amqp.EXCHANGE_DELETE):
			fmt.Printf("[DEBUG] Received exchange.delete request: %+v\n", request)
			channelId := request.Channel
			fmt.Printf("[DEBUG] Channel: %d\n", channelId)
			content, ok := request.Content.(*amqp.ExchangeDeleteMessage)
			if !ok {
				fmt.Printf("Invalid content type for ExchangeDeclareMessage")
				return nil, fmt.Errorf("Invalid content type for ExchangeDeclareMessage")
			}
			fmt.Printf("[DEBUG] Content: %+v\n", content)
			exchangeName := content.ExchangeName
			// ifUnused := content.IfUnused
			// noWait := content.NoWait

			vh := b.VHosts["/"]
			err := vh.DeleteExchange(exchangeName)
			if err != nil {
				return nil, err
			}

			frame := amqp.ResponseMethodMessage{
				Channel:  channelId,
				ClassID:  request.ClassID,
				MethodID: uint16(amqp.EXCHANGE_DELETE_OK),
				Content:  amqp.ContentList{},
			}.FormatMethodFrame()

			shared.SendFrame(conn, frame)
			return nil, nil

		default:
			return nil, fmt.Errorf("unsupported command")
		}

	case uint16(amqp.QUEUE):
		switch request.MethodID {
		case uint16(amqp.QUEUE_DECLARE):
			fmt.Printf("[DEBUG] Received queue declare request: %+v\n", request)
			channelId := request.Channel
			content, ok := request.Content.(*amqp.QueueDeclareMessage)
			if !ok {
				fmt.Printf("Invalid content type for ExchangeDeclareMessage")
				return nil, fmt.Errorf("Invalid content type for ExchangeDeclareMessage")
			}
			fmt.Printf("[DEBUG] Content: %+v\n", content)
			queueName := content.QueueName

			vh := b.VHosts["/"]

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
				Channel:  channelId,
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
			shared.SendFrame(conn, frame)
			return nil, nil

		case uint16(amqp.QUEUE_BIND):
			fmt.Printf("[DEBUG] Received queue bind request: %+v\n", request)
			channelId := request.Channel
			content, ok := request.Content.(*amqp.QueueBindMessage)
			if !ok {
				fmt.Printf("Invalid content type for ExchangeDeclareMessage")
				return nil, fmt.Errorf("Invalid content type for ExchangeDeclareMessage")
			}
			fmt.Printf("[DEBUG] Content: %+v\n", content)
			vh := b.VHosts["/"]
			queue := content.Queue
			exchange := content.Exchange
			routingKey := content.RoutingKey
			// noWait := content.NoWait

			err := vh.BindQueue(exchange, queue, routingKey)
			if err != nil {
				fmt.Printf("[DEBUG] Error binding to default exchange: %v\n", err)
				return nil, err
			}
			frame := amqp.ResponseMethodMessage{
				Channel:  channelId,
				ClassID:  request.ClassID,
				MethodID: uint16(amqp.QUEUE_BIND_OK),
				Content:  amqp.ContentList{},
			}.FormatMethodFrame()
			shared.SendFrame(conn, frame)
			return nil, nil

		case uint16(amqp.QUEUE_DELETE):
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// queueName := parts[1]
			// err := b.deleteQueue(queueName)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s deleted", queueName)}, nil

		case uint16(amqp.QUEUE_UNBIND):
			// if len(parts) != 4 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// exchangeName := parts[1]
			// queueName := parts[2]
			// routingKey := parts[3]
			// b.DeletBinding(exchangeName, queueName, routingKey)
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Binding deleted")}, nil
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
				return nil, fmt.Errorf("Channel not found")
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
				fmt.Printf("[DEBUG] All fields shall be filled -> current state: %+v\n", currentState)
				if len(currentState.Body) != int(currentState.BodySize) {
					fmt.Printf("[DEBUG] Body size is not correct: %d != %d\n", len(currentState.Body), currentState.BodySize)
					return nil, fmt.Errorf("body size is not correct: %d != %d", len(currentState.Body), currentState.BodySize)
				}
				publishRequest := currentState.MethodFrame.Content.(*amqp.BasicPublishMessage)
				exchanege := publishRequest.Exchange
				routingKey := publishRequest.RoutingKey
				body := currentState.Body
				props := currentState.HeaderFrame.Properties
				v := b.VHosts["/"]
				// v.Publish(exchanege, routingKey, body, props)
				_, err := v.Publish(exchanege, routingKey, body, props)
				if err == nil {
					log.Printf("[DEBUG] Published message to exchange=%s, routingKey=%s, body=%s", exchanege, routingKey, string(body))
					b.Connections[conn].Channels[channel] = &amqp.ChannelState{}
				}
				return nil, err

			}
		case uint16(amqp.BASIC_GET):
			vhost := b.VHosts["/"] // TODO: set the selected vhost
			getMsg := request.Content.(*amqp.BasicGetMessage)
			queue := getMsg.Queue
			channelId := request.Channel
			messageCount, err := vhost.GetMessageCount(queue)
			if err != nil {
				fmt.Println("Error getting message count:", err)
				return nil, err
			}
			msg := vhost.GetMessage(queue)
			if msg == nil {
				// Send Basic.GetEmpty
				// I figure out that expects an octet here
				reserved1 := amqp.KeyValue{
					Key:   amqp.INT_OCTET,
					Value: getMsg.Reserved1,
				}
				frame := amqp.ResponseMethodMessage{
					Channel:  channelId,
					ClassID:  request.ClassID,
					MethodID: uint16(amqp.BASIC_GET_EMPTY),
					Content:  amqp.ContentList{KeyValuePairs: []amqp.KeyValue{reserved1}},
				}.FormatMethodFrame()
				fmt.Printf("[DEBUG] Sending get-enpty frame: %x\n", frame)
				shared.SendFrame(conn, frame)
				return nil, nil
			}

			// Send Basic.GetOk + header + body
			msgGetOk := &amqp.BasicGetOk{
				DeliveryTag:  1,
				Redelivered:  false,
				Exchange:     msg.Exchange,
				RoutingKey:   msg.RoutingKey,
				MessageCount: uint32(messageCount),
			}

			frame := amqp.ResponseMethodMessage{
				Channel:  channelId,
				ClassID:  request.ClassID,
				MethodID: uint16(amqp.BASIC_GET_OK),
				Content:  *amqp.EncodeGetOkToContentList(msgGetOk),
			}.FormatMethodFrame()

			err = shared.SendFrame(conn, frame)
			log.Printf("[DEBUG] Sent message from queue %s: ID=%s", queue, msg.ID)

			if err != nil {
				fmt.Printf("[DEBUG] Error sending frame: %v\n", err)
				return nil, err
			}

			responseContent := amqp.ResponseContent{
				Channel: channelId,
				ClassID: request.ClassID,
				Weight:  0,
				Message: *msg,
			}
			// Header
			frame = responseContent.FormatHeaderFrame()
			shared.SendFrame(conn, frame)
			// Body
			frame = responseContent.FormatBodyFrame()
			shared.SendFrame(conn, frame)
			return nil, nil

		case uint16(amqp.BASIC_ACK):
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// msgID := parts[1]
			// // get queue from consumerID
			// consumer, ok := b.Consumers[consumerID]
			// if !ok {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Consumer not found"}, nil
			// }
			// queue := consumer.Queue
			// b.acknowledge(queue, consumerID, msgID)
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Message ID %s acknowledged", msgID)}, nil

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

func (b *Broker) GetVHostFromName(vhostName string) *vhost.VHost {
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
