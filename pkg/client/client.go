package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient return websocket client connection
type WebSocketClient struct {
	configStr string

	mu     sync.RWMutex
	wsconn *websocket.Conn
}

// NewWebSocketClient create new websocket connection
func NewWebSocketClient(host, channel string) (*WebSocketClient, error) {
	conn := WebSocketClient{}

	u := url.URL{Scheme: "ws", Host: host, Path: channel}
	conn.configStr = u.String()

	go conn.Connect()
	go conn.listen()
	return &conn, nil
}

func (conn *WebSocketClient) Connect() *websocket.Conn {
	conn.mu.RLock()
	if conn.wsconn != nil {
		defer conn.mu.RUnlock()
		return conn.wsconn
	}
	conn.mu.RUnlock()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		ws, _, err := websocket.DefaultDialer.Dial(conn.configStr, nil)
		if err != nil {
			conn.log("connect", err, fmt.Sprintf("Cannot connect to websocket: %s", conn.configStr))
			continue
		}
		conn.log("connect", nil, fmt.Sprintf("connected to websocket to %s", conn.configStr))
		conn.mu.Lock()
		conn.wsconn = ws
		conn.mu.Unlock()
		return conn.wsconn
	}
}

func (conn *WebSocketClient) listen() {
	conn.log("listen", nil, fmt.Sprintf("listen for the messages: %s", conn.configStr))
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		ws := conn.Connect()
		if ws == nil {
			return
		}
		_, bytMsg, err := ws.ReadMessage()
		if err != nil {
			conn.log("listen", err, "Cannot read websocket message")
			conn.Stop()
			break
		}
		conn.log("listen", nil, fmt.Sprintf("receive msg %s\n", bytMsg))
	}
}

// Write data to the websocket server or drop it after 50ms
func (conn *WebSocketClient) Write(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	conn.mu.Lock()
	defer conn.mu.Unlock()
	ws := conn.wsconn
	if ws == nil {
		err := fmt.Errorf("conn.ws is nil")
		return err
	}

	if err := ws.WriteMessage(
		websocket.TextMessage,
		data,
	); err != nil {
		conn.log("Write", nil, "WebSocket Write Error")
	}
	conn.log("Write", nil, fmt.Sprintf("sent: %s", data))
	return nil
}

// Close will send close message and shutdown websocket connection
func (conn *WebSocketClient) Stop() {
	conn.mu.Lock()
	if conn.wsconn != nil {
		conn.wsconn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.wsconn.Close()
		conn.wsconn = nil
	}
	conn.mu.Unlock()
}

// Log print log statement
// In real word I would recommend to use zerolog or any other solution
func (conn *WebSocketClient) log(f string, err error, msg string) {
	if err != nil {
		fmt.Printf("Error in func: %s, err: %v, msg: %s\n", f, err, msg)
	} else {
		fmt.Printf("Log in func: %s, %s\n", f, msg)
	}
}
