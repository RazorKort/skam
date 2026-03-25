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
	Url       string
	Conn      *websocket.Conn
	SendChan  chan []byte
	MsgChan   chan Message
	DoneChan  chan struct{}
	mu        sync.Mutex
	closeOnce sync.Once
	reconnect bool
	chanMu    sync.RWMutex // защита замены каналов
}

func (c *Client) NewWSClient() (*WSClient, error) {
	url := "wss://" + c.ServerUrl + "/ws?token=" + c.token

	ws := &WSClient{
		Url:       url,
		SendChan:  make(chan []byte, 10),
		MsgChan:   make(chan Message, 20),
		DoneChan:  make(chan struct{}),
		reconnect: true,
	}
	err := ws.connect()
	if err != nil {
		return nil, err
	}
	go ws.reconnectLoop()
	return ws, nil
}

func (ws *WSClient) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(ws.Url, nil)
	if err != nil {
		return err
	}
	ws.mu.Lock()
	ws.Conn = conn
	ws.mu.Unlock()

	// Создаём новые каналы
	newSendChan := make(chan []byte, 10)
	newMsgChan := make(chan Message, 20)
	newDoneChan := make(chan struct{})

	// Заменяем каналы под блокировкой
	ws.chanMu.Lock()
	oldSendChan := ws.SendChan
	oldMsgChan := ws.MsgChan
	oldDoneChan := ws.DoneChan
	ws.SendChan = newSendChan
	ws.MsgChan = newMsgChan
	ws.DoneChan = newDoneChan
	ws.chanMu.Unlock()

	// Закрываем старые каналы, чтобы старые горутины завершились
	if oldDoneChan != nil {
		close(oldDoneChan)
	}
	if oldSendChan != nil {
		close(oldSendChan)
	}
	if oldMsgChan != nil {
		close(oldMsgChan)
	}

	// Запускаем новые горутины
	go ws.readLoop()
	go ws.writeLoop()
	return nil
}

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
				return
			}
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				log.Println("ws read error: ", err)
				ws.Close()
				return
			}
			var msg Message
			err = json.Unmarshal(msgBytes, &msg)
			if err != nil {
				log.Println("ws unmarshal error: ", err)
				continue
			}
			ws.chanMu.RLock()
			msgChan := ws.MsgChan
			ws.chanMu.RUnlock()
			select {
			case msgChan <- msg:
			case <-time.After(time.Second):
				log.Println("ws MsgChan full, dropping message")
			}
		}
	}
}

func (ws *WSClient) writeLoop() {
	for {
		select {
		case <-ws.DoneChan:
			return
		case msg, ok := <-ws.SendChan:
			if !ok {
				return
			}
			ws.mu.Lock()
			conn := ws.Conn
			ws.mu.Unlock()
			if conn == nil {
				return
			}
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("ws write error: ", err)
				ws.Close()
				return
			}
		}
	}
}

func (ws *WSClient) Close() {
	ws.closeOnce.Do(func() {
		ws.reconnect = false
		ws.chanMu.Lock()
		if ws.DoneChan != nil {
			close(ws.DoneChan)
		}
		ws.chanMu.Unlock()
		ws.mu.Lock()
		if ws.Conn != nil {
			ws.Conn.Close()
		}
		ws.mu.Unlock()
		// Каналы SendChan и MsgChan не закрываем здесь, они закрываются при замене
	})
}

func (ws *WSClient) reconnectLoop() {
	for ws.reconnect {
		<-ws.DoneChan

		for ws.reconnect {
			log.Println("Attempting to reconnect...")
			err := ws.connect()
			if err == nil {
				log.Println("Reconnected successfully")
				break
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func (ws *WSClient) SendMessage(msg Message) error {
	sendMsg := MessageToSend{
		Receiver_id: msg.Receiver_id,
		Message:     msg.Message,
		Sender_id:   msg.Sender_id,
		Created_at:  msg.Created_at,
	}
	data, err := json.Marshal(sendMsg)
	if err != nil {
		return err
	}

	ws.chanMu.RLock()
	sendChan := ws.SendChan
	ws.chanMu.RUnlock()
	if sendChan == nil {
		return fmt.Errorf("websocket not ready")
	}

	select {
	case sendChan <- data:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send timeout")
	}
}
