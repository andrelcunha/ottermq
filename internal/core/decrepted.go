package core

// func (b *VHost) handleConnection(conn net.Conn) {
// 	defer func() {
// 		conn.Close()
// 		b.cleanupConnection(conn)
// 	}()
// 	reader := bufio.NewReader(conn)

// 	b.registerConnection(conn)
// 	for isAuthenticated := false; !isAuthenticated; {
// 		msg, err := reader.ReadString('\n')
// 		if err != nil {
// 			log.Println("Failed to read message:", err)
// 			return
// 		}
// 		msg = strings.TrimSpace(msg)
// 		parts := strings.Fields(msg)
// 		if len(parts) != 3 || parts[0] != "AUTH" {
// 			conn.Write([]byte("Invalid handshake\n"))
// 			continue
// 		}
// 		username, password := parts[1], parts[2]
// 		isAuthenticated = b.authenticate(username, password)
// 		var response common.CommandResponse
// 		if isAuthenticated {
// 			response = common.CommandResponse{
// 				Status:  "OK",
// 				Message: fmt.Sprintf("User '%s' authenticated successfully", username),
// 			}
// 		} else {
// 			response = common.CommandResponse{
// 				Status:  "ERROR",
// 				Message: "Invalid credentials",
// 			}
// 		}
// 		responseJSON, err := json.Marshal(response)
// 		if err != nil {
// 			log.Println("Failed to marshal response:", err)
// 			return
// 		}
// 		conn.Write(append(responseJSON, '\n'))
// 	}

// 	consumerID := b.registerSessionAndConsummer(conn)

// 	go b.sendHeartbeat(conn)

// 	// reader := bufio.NewReader(conn)
// 	for {
// 		conn.SetReadDeadline(time.Now().Add(b.HeartbeatInterval * 2))
// 		msg, err := reader.ReadString('\n')
// 		if err != nil {
// 			if err == io.EOF {
// 				log.Println("Connection closed by client")
// 			} else {
// 				log.Println("Connection closed or heartbeat timeout: ", err)
// 			}
// 			break
// 		}

// 		if strings.TrimSpace(msg) == "HEARTBEAT" {
// 			b.mu.Lock()
// 			b.LastHeartbeat[conn] = time.Now()
// 			b.mu.Unlock()
// 			// log.Println("Received heartbeat")
// 			continue
// 		}

// 		log.Printf("Received: %s\n", msg)
// 		response, err := b.processCommand(msg, consumerID)
// 		if err != nil {
// 			log.Println("ERROR: ", err)
// 			response = common.CommandResponse{
// 				Status:  "error",
// 				Message: err.Error(),
// 			}
// 		}

// 		responseJSON, err := json.Marshal(response)
// 		if err != nil {
// 			log.Println("Failed to serialize response:", err)
// 			continue
// 		}
// 		_, err = conn.Write(append(responseJSON, '\n'))
// 		if err != nil {
// 			log.Println("Failed to write response:", err)
// 			break
// 		}
// 	}
// }
