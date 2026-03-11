package back

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	Url       string
	Conn      *websocket.Conn
	SendChan  chan []byte
	MsgChan   chan Message
	DoneChan  chan struct{}
	mu        sync.Mutex
	closeOnce sync.Once
}

// я сам не понял че я тут написал. но вроде работает
func NewWSClient(c *Client) (*WSClient, error) {
	url := "wss://" + c.ServerUrl + "/ws?token=" + c.token

	ws := &WSClient{
		Url:      url,
		SendChan: make(chan []byte, 10),
		MsgChan:  make(chan Message, 20),
		DoneChan: make(chan struct{}),
	}
	err := ws.connect()
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (ws *WSClient) readLoop() {
	for {
		_, msgBytes, err := ws.Conn.ReadMessage()
		if err != nil {
			log.Println("ws read error: ", err)
			ws.Close()
			return
		}
		var msg Message
		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			log.Println("ws unmarshal error : ", err)
			continue
		}
		ws.MsgChan <- msg
	}
}

func (ws *WSClient) writeLoop() {
	for msg := range ws.SendChan {
		ws.mu.Lock()
		err := ws.Conn.WriteMessage(websocket.TextMessage, msg)
		ws.mu.Unlock()
		if err != nil {
			log.Println("ws write error : ", err)
			ws.Close()
			return
		}
	}
}

func (ws *WSClient) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(ws.Url, nil)
	if err != nil {
		return err
	}
	ws.Conn = conn

	go ws.readLoop()
	go ws.writeLoop()

	return nil
}

func (ws *WSClient) Close() {
	ws.closeOnce.Do(func() {
		ws.Conn.Close()
		close(ws.DoneChan)
	})
}

func (ws *WSClient) reconnectLoop() {
	for {
		<-ws.DoneChan

		for {
			err := ws.connect()
			if err == nil {
				ws.DoneChan = make(chan struct{})
				break
			}
			time.Sleep(3 * time.Second)
		}
	}
}
