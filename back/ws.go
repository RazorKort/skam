package back

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	Url          string
	Conn         *websocket.Conn
	SendChan     chan []byte
	MsgChan      chan Message
	MsgSendChan  chan MessageToSend
	DoneChan     chan struct{}
	mu           sync.Mutex
	reconnecting bool
	stop         bool
}

// 100% vibe coded
// connect establishes a new connection, closing the old one.
func (ws *WSClient) connect() error {
	ws.mu.Lock()
	oldConn := ws.Conn
	ws.mu.Unlock()
	if oldConn != nil {
		oldConn.Close()
	}

	conn, _, err := websocket.DefaultDialer.Dial(ws.Url, nil)
	if err != nil {
		return err
	}

	ws.mu.Lock()
	ws.Conn = conn
	ws.mu.Unlock()
	return nil
}

// reconnect attempts to re-establish the connection with exponential backoff.
func (ws *WSClient) reconnect() {
	ws.mu.Lock()
	if ws.stop || ws.reconnecting {
		ws.mu.Unlock()
		return
	}
	ws.reconnecting = true
	ws.mu.Unlock()
	defer func() {
		ws.mu.Lock()
		ws.reconnecting = false
		ws.mu.Unlock()
	}()

	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		if ws.stop {
			return
		}
		log.Println("Attempting to reconnect...")
		if err := ws.connect(); err == nil {
			log.Println("Reconnected successfully")

			// Reset DoneChan to a new channel (the old one is already closed)
			ws.mu.Lock()
			if ws.DoneChan != nil {
				close(ws.DoneChan) // just to be safe
			}
			ws.DoneChan = make(chan struct{})
			ws.mu.Unlock()

			// Restart the read and write loops
			go ws.readLoop()
			go ws.writeLoop()
			return
		}
		log.Printf("Reconnect failed, retrying in %v...", backoff)
		select {
		case <-time.After(backoff):
		}
		if backoff < maxBackoff {
			backoff *= 2
		}
	}
}

// readLoop reads messages from the WebSocket.
func (ws *WSClient) readLoop() {
	for {
		select {
		case <-ws.DoneChan:
			return
		default:
			ws.mu.Lock()
			conn := ws.Conn
			ws.mu.Unlock()
			if conn == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				log.Println("ws read error:", err)
				// Trigger reconnection and exit this goroutine
				go ws.reconnect()
				return
			}
			var msg Message
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("ws unmarshal error:", err)
				continue
			}
			select {
			case ws.MsgChan <- msg:
			case <-time.After(time.Second):
				log.Println("ws MsgChan full, dropping message")
			}
		}
	}
}

// writeLoop sends data from SendChan to the WebSocket.
func (ws *WSClient) writeLoop() {
	for {
		select {
		case <-ws.DoneChan:
			return
		case data, ok := <-ws.SendChan:
			if !ok {
				return
			}
			ws.mu.Lock()
			conn := ws.Conn
			ws.mu.Unlock()
			if conn == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("ws write error:", err)
				// Trigger reconnection and exit this goroutine
				go ws.reconnect()
				return
			}
		}
	}
}

// sendWorker processes outgoing messages.
func (ws *WSClient) sendWorker() {
	for msg := range ws.MsgSendChan {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			continue
		}
		select {
		case ws.SendChan <- data:
		case <-time.After(5 * time.Second):
			log.Printf("send timeout")
		}
	}
}

// Close gracefully shuts down the current connection.
func (ws *WSClient) Close() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.DoneChan != nil {
		close(ws.DoneChan)
		ws.DoneChan = nil
	}
	if ws.Conn != nil {
		ws.Conn.Close()
		ws.Conn = nil
	}
}

// Stop permanently stops the client.
func (ws *WSClient) Stop() {
	ws.mu.Lock()
	ws.stop = true
	ws.mu.Unlock()
	ws.Close()
}

// NewWSClient creates a new WebSocket client and starts its goroutines.
func (c *Client) NewWSClient() (*WSClient, error) {
	url := "wss://" + c.ServerUrl + "/ws?token=" + c.token
	ws := &WSClient{
		Url:         url,
		SendChan:    make(chan []byte, 10),
		MsgChan:     make(chan Message, 20),
		MsgSendChan: make(chan MessageToSend, 20),
		DoneChan:    make(chan struct{}),
	}
	if err := ws.connect(); err != nil {
		return nil, err
	}
	go ws.readLoop()
	go ws.writeLoop()
	go ws.sendWorker()
	return ws, nil
}

// SendMessage queues a message for sending.
func (ws *WSClient) SendMessage(msg Message) error {
	sendMsg := MessageToSend{
		Receiver_id: msg.Receiver_id,
		Message:     msg.Message,
		Sender_id:   msg.Sender_id,
		Tmp_id:      msg.Id,
	}
	select {
	case ws.MsgSendChan <- sendMsg:
		return nil
	case <-time.After(1 * time.Second):
		return fmt.Errorf("outgoing queue full")
	}
}
