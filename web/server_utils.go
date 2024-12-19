package web

import (
	"fmt"
	"log"
	"net"
	"time"
)

func getConnection(brokerAddr string) (net.Conn, error) {
	var conn net.Conn
	err := error(nil)
	tries := 1
	for {
		if tries > 3 {
			return nil, fmt.Errorf("Broker did not respond after 3 retries")
		}
		conn, err = net.Dial("tcp", brokerAddr)
		if err == nil {
			break
		}
		wait := 3 * tries
		log.Printf("Broker is not ready. Retrying in %d seconds...\n", wait)
		time.Sleep(time.Duration(wait) * time.Second)
		tries++
	}
	return conn, nil
}
