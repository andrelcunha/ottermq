package client

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/constants/tx"
	"github.com/andrelcunha/ottermq/pkg/connection/shared"
)

var (
	version = "0.6.0-alpha"
)

const (
	platform = "golang"
	product  = "OtterMQ Client"
)

type Client struct {
	conn     net.Conn
	config   *shared.ClientConfig
	channels map[uint16]*Channel
	Mu       sync.Mutex
	cmdCh    chan []byte
	respCh   chan interface{}
	closeCh  chan struct{}
}

func NewClient(config *shared.ClientConfig) *Client {
	return &Client{
		config:   config,
		channels: make(map[uint16]*Channel),
		cmdCh:    make(chan []byte),
		respCh:   make(chan interface{}),
		closeCh:  make(chan struct{}),
	}
}

func (c *Client) Dial(host, port string) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("ERROR: %s\n", err.Error())
	}
	c.conn = conn
	defer conn.Close()

	capabilities := map[string]interface{}{
		"basic.nack":             true,
		"connection.blocked":     true,
		"consumer_cancel_notify": true,
		"publisher_confirms":     true,
	}

	clientProperties := map[string]interface{}{
		"capabilities": capabilities,
		"product":      product,
		"version":      version,
		"platform":     platform,
	}
	configurations := map[string]interface{}{
		"username":         c.config.Username,
		"password":         c.config.Password,
		"vhost":            c.config.Vhost,
		"clientProperties": clientProperties,
		"mechanism":        "PLAIN",
		"locale":           "en_US",
	}
	if err := clientHandshake(&configurations, conn); err != nil {
		return err
	}

	go c.sendHeartbeats()
	go c.receiveLoop(conn)

	for {
		select {
		case frame := <-c.cmdCh:
			err := shared.SendFrame(c.conn, frame)
			if err != nil {
				fmt.Printf("Failed to open channel: %v", err)

			}
		case <-c.closeCh:
			return nil
		}
	}
}

func (c *Client) receiveLoop(conn net.Conn) {
	fmt.Println("Starting receive loop")
	for {
		frame, err := shared.ReadFrame(conn)
		fmt.Printf("Received frame: %x\n", frame)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Connection timeout: %v", err)
			}
			if err == io.EOF {
				log.Printf("Connection closed by server: %v", conn.RemoteAddr())
				return
			}
			log.Printf("Error reading frame: %v", err)
			return
		}
		log.Printf("[DEBUG] Received frame: %x", frame)

		response, err := shared.ParseFrame(nil, frame)
		if err != nil {
			log.Printf("Error parsing frame: %v", err)
			// return
		}
		if response != nil {
			// c.respCh <- response

		}
	}
}

// Client sends ProtocolHeader
// Server responds with connection.start
// Client responds with connection.start-ok
// Server responds with connection.tune
// Client responds with connection.tune-ok
// Client sends connection.open
// Server responds with connection.open-ok
func clientHandshake(configurations *map[string]interface{}, conn net.Conn) error {
	log.Printf("Starting connection handshake")
	// Send Protocol Header
	if err := shared.SendProtocolHeader(conn); err != nil {
		log.Fatalf("Failed to send protocol header: %v", err)
	}

	/** connection.start **/
	// Read connection.start frame
	frame, err := shared.ReadFrame(conn)
	if err != nil {
		err = fmt.Errorf("Failed to read connection.start frame: %v", err)
		return err
	}
	if isAMQPVersionResponse(frame) {
		log.Printf("Received AMQP version response: %s", frame)
		return nil
	}

	//TODO: Refactor - remove processing from ParseFrame
	// This processing is concerned to the client only.
	untypedResp, err := shared.ParseFrame(configurations, frame)
	if err != nil {
		err = fmt.Errorf("Failed to parse connection.start frame: %v", err)
		return err
	}
	startResponse, ok := untypedResp.(*shared.ConnectionStartFrame)
	if !ok {
		err = fmt.Errorf("Type assertion failed")
		return err
	}

	startOkRequest, err := shared.CreateConnectionStartOkPayload(configurations, startResponse)
	if err != nil {
		err = fmt.Errorf("Failed to create connection.start-ok frame: %v", err)
		return err
	}
	startOkFrame := shared.CreateConnectionStartOkFrame(&startOkRequest)
	if err := shared.SendFrame(conn, startOkFrame); err != nil {
		err = fmt.Errorf("Failed to send connection.start-ok frame: %v", err)
		return err
	}

	/** connection.tune **/
	// Read connection.tune frame
	frame, err = shared.ReadFrame(conn)
	if err != nil {
		err = fmt.Errorf("Failed to read connection.tune frame: %v", err)
		return err
	}
	untypedResp, err = shared.ParseFrame(configurations, frame)
	tuneOkResponse, ok := untypedResp.([]byte)
	if !ok {
		err = fmt.Errorf("Type assertion ConnectionTuneOkFrame failed")
		return err
	}
	if err := shared.SendFrame(conn, tuneOkResponse); err != nil {
		log.Fatalf("Failed to send connection.tune-ok frame: %v", err)
	}
	/** end of connection.tune **/

	/** connection.open **/
	vhost := (*configurations)["vhost"].(string)
	openOkFrame := shared.CreateConnectionOpenFrame(vhost)
	if err := shared.SendFrame(conn, openOkFrame); err != nil {
		log.Fatalf("Failed to send connection.open-ok frame: %v", err)
	}

	frame, err = shared.ReadFrame(conn)
	if err != nil {
		log.Fatalf("Failed to read frame: %v", err)
	}
	fmt.Printf("Received connection.open-ok: %x\n", frame)
	_, err = shared.ParseFrame(configurations, frame)

	// Parse open-ok frame

	fmt.Println("AMQP handshake complete")
	return nil
}

func isAMQPVersionResponse(frame []byte) bool {
	return bytes.Equal(frame[:4], []byte("AMQP"))
}

func (c *Client) sendHeartbeats() {
	// heartbeatInterval := 5 // You can set this to the appropriate interval
	heartbeatInterval := c.config.HeartbeatInterval
	ticker := time.NewTicker(time.Duration(heartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			heartbeatFrame := shared.CreateHeartbeatFrame()
			err := shared.SendFrame(c.conn, heartbeatFrame)
			if err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				return
			}
			log.Println("[DEBUG] Heartbeat sent")
		}
	}
}

func (c *Client) nextChannelID() uint16 {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	// get the biggest channel ID
	maxID := uint16(0)
	for id := range c.channels {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

func (c *Client) Channel() (*Channel, error) {
	channelId := c.nextChannelID()
	channel := &Channel{
		Id: channelId,
		c:  c,
	}
	c.Mu.Lock()
	c.channels[channelId] = channel
	c.Mu.Unlock()

	frame := amqp.ResponseMethodMessage{
		Channel:  channelId,
		ClassID:  uint16(constants.CHANNEL),
		MethodID: uint16(constants.CHANNEL_OPEN),
		Content: amqp.ContentList{
			KeyValuePairs: []amqp.KeyValue{
				{
					Key:   amqp.INT_OCTET,
					Value: uint8(0),
				},
			},
		},
	}.FormatMethodFrame()
	c.PostFrame(frame)

	response := <-c.respCh
	responseFrame, ok := response.(*amqp.RequestMethodMessage)
	if !ok {
		return nil, fmt.Errorf("Failed to parse response frame")
	}
	if responseFrame.ClassID != uint16(constants.CHANNEL) || responseFrame.MethodID != uint16(constants.CHANNEL_OPEN_OK) {
		return nil, fmt.Errorf("Failed to open channel")
	}

	return channel, nil
}

func (c *Client) Shutdown() {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	for _, channel := range c.channels {
		channel.Close()
	}

	//create a connection.close frame
	frame := amqp.ResponseMethodMessage{
		Channel:  c.nextChannelID(),
		ClassID:  uint16(constants.CONNECTION),
		MethodID: uint16(constants.CONNECTION_CLOSE),
		Content: amqp.ContentList{
			KeyValuePairs: []amqp.KeyValue{
				{
					Key:   amqp.INT_OCTET,
					Value: uint8(0),
				},
			},
		},
	}.FormatMethodFrame()
	c.PostFrame(frame)
	response := <-c.respCh
	responseFrame, ok := response.(*amqp.RequestMethodMessage)
	if !ok {
		fmt.Printf("Failed to parse response frame")
	}
	if responseFrame.ClassID != uint16(constants.CHANNEL) || responseFrame.MethodID != uint16(constants.CHANNEL_CLOSE_OK) {
		fmt.Printf("Failed to open channel")
	}

	c.closeCh <- struct{}{}
	c.conn.Close()
}

func (c *Client) PostFrame(frame []byte) {
	c.cmdCh <- frame
}

func (c *Client) processRequest(conn net.Conn, request *amqp.RequestMethodMessage) (interface{}, error) {
	switch request.ClassID {

	case uint16(constants.CONNECTION):
		switch request.MethodID {

		case uint16(constants.CONNECTION_CLOSE):
			c.cleanupConnection()
			frame := amqp.ResponseMethodMessage{
				Channel:  request.Channel,
				ClassID:  uint16(constants.CONNECTION),
				MethodID: uint16(constants.CONNECTION_CLOSE_OK),
				Content:  amqp.ContentList{},
			}.FormatMethodFrame()
			shared.SendFrame(conn, frame)
			return nil, nil
		default:
			log.Printf("[DEBUG] Unknown connection method: %d", request.MethodID)
			return nil, fmt.Errorf("Unknown connection method: %d", request.MethodID)
		}

	case uint16(constants.CHANNEL):
		switch request.MethodID {
		// case uint16(constants.CHANNEL_OPEN_OK):
		// fmt.Printf("[DEBUG] Received channel open request: %+v\n", request)
		// channelId := request.Channel
		// c.Mu.Lock()
		// // Check if the channel is already open

		// c.Mu.Unlock()
		// frame := amqp.ResponseMethodMessage{
		// 	Channel:  channelId,
		// 	ClassID:  request.ClassID,
		// 	MethodID: uint16(constants.CHANNEL_OPEN_OK),
		// 	Content: amqp.ContentList{
		// 		KeyValuePairs: []amqp.KeyValue{
		// 			{
		// 				Key:   amqp.INT_OCTET,
		// 				Value: uint8(0),
		// 			},
		// 		},
		// 	},
		// }.FormatMethodFrame()
		// shared.SendFrame(conn, frame)
		// fmt.Printf("[DEBUG] sending frame: %x\n", frame)
		// return nil, nil

		case uint16(constants.CHANNEL_CLOSE_OK):
			channelId := request.Channel
			c.Mu.Lock()
			delete(c.channels, channelId)
			c.Mu.Unlock()
			return nil, nil

		default:
			log.Printf("[DEBUG] Unknown channel method: %d", request.MethodID)
			return nil, fmt.Errorf("Unknown channel method: %d", request.MethodID)
		}

		// Handle channel-related commands
	case uint16(constants.EXCHANGE):
		switch request.MethodID {
		case uint16(constants.EXCHANGE_DECLARE):
			// content := request.Content.()

			// err := b.createExchange(exchangeName, ExchangeType(typ))
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Exchange %s of type %s created", exchangeName, typ)}, nil
		case uint16(constants.EXCHANGE_DELETE):
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// exchangeName := parts[1]
			// err := b.deleteExchange(exchangeName)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Exchange %s deleted", exchangeName)}, nil

		default:
			return nil, fmt.Errorf("unsupported command")
		}
		// Handle exchange-related commands

	case uint16(constants.QUEUE):
		switch request.MethodID {
		case uint16(constants.QUEUE_DECLARE):
			// Handle queue declaration
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// queueName := parts[1]
			// _, err := b.createQueue(queueName)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// // Bind the queue to the default exchange with the same name as the queue
			// err = b.bindToDefaultExchange(queueName)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s created", queueName)}, nil
		case uint16(constants.QUEUE_BIND):
			// if len(parts) < 3 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// exchangeName := parts[1]
			// queueName := parts[2]
			// routingKey := ""
			// if len(parts) == 4 {
			// 	routingKey = parts[3]
			// }
			// err := b.bindQueue(exchangeName, queueName, routingKey)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s bound to exchange %s", queueName, exchangeName)}, err
		case uint16(constants.QUEUE_DELETE):
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// queueName := parts[1]
			// err := b.deleteQueue(queueName)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s deleted", queueName)}, nil

		case uint16(constants.QUEUE_UNBIND):
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
	case uint16(constants.BASIC):
		switch request.MethodID {
		case uint16(constants.BASIC_QOS):
		case uint16(constants.BASIC_CONSUME):
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// queueName := parts[1]
			// fmt.Println("Consuming from queue:", queueName)
			// msg := b.consume(queueName, consumerID)
			// if msg == nil {
			// 	return common.CommandResponse{Status: "OK", Message: "No messages available", Data: ""}, nil
			// }
			// return common.CommandResponse{Status: "OK", Data: msg}, nil
		case uint16(constants.BASIC_CANCEL):
		case uint16(constants.BASIC_PUBLISH):
			// if len(parts) < 4 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// exchangeName := parts[1]
			// routingKey := parts[2]
			// message := strings.Join(parts[3:], " ")
			// msgId, err := b.publish(exchangeName, routingKey, message)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// var data struct {
			// 	MessageID string `json:"message_id"`
			// }
			// data.MessageID = msgId
			// return common.CommandResponse{Status: "OK", Message: "Message sent", Data: data}, nil
		case uint16(constants.BASIC_RETURN):
		case uint16(constants.BASIC_DELIVER):
		case uint16(constants.BASIC_GET):
		case uint16(constants.BASIC_GET_EMPTY):
			// Handle message retrieval
		case uint16(constants.BASIC_ACK):
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

		case uint16(constants.BASIC_REJECT):
		case uint16(constants.BASIC_RECOVER_ASYNC):
		case uint16(constants.BASIC_RECOVER):
		case uint16(constants.BASIC_RECOVER_OK):
		default:
			return nil, fmt.Errorf("unsupported command")
		}
	case uint16(constants.TX):
		// Handle transaction-related commands
		switch request.MethodID {
		case uint16(tx.SELECT):
			// Handle transaction selection
		case uint16(tx.COMMIT):
			// Handle transaction commit
		case uint16(tx.ROLLBACK):
			// Handle transaction rollback
		default:
			return nil, fmt.Errorf("unsupported command")
		}
	default:
		return nil, fmt.Errorf("unsupported command")
	}
	return nil, nil
}

func (c *Client) cleanupConnection() {
	log.Println("Cleaning connection")
	c.Mu.Lock()
	c.conn.Close()
	c.Mu.Unlock()
	for _, ch := range c.channels {
		delete(c.channels, ch.Id)
	}
}
