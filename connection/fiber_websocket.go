package connection

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber"
	"github.com/gofiber/websocket"
	"github.com/google/uuid"
)

/*

Websocket connection and connection Factory

wraps the websocket connection and forwards it to manager.

*/

//WebSocketConnection - a web socket connection struct
type WebSocketConnection struct {
	conn       *websocket.Conn
	closed     bool
	listeners  map[string]func(Connection, string, string)
	listenSync sync.Mutex
	uuid       uuid.UUID
}

//Listen - add listener for messages
func (conn *WebSocketConnection) Listen(listener func(Connection, string, string)) string {
	conn.listenSync.Lock()
	defer conn.listenSync.Unlock()
	if conn.listeners == nil {
		conn.listeners = make(map[string]func(Connection, string, string))
	}
	uid := uuid.New().String()
	for _, ok := conn.listeners[uid]; ok; {
		uid = uuid.New().String()
		_, ok = conn.listeners[uid]
	}
	conn.listeners[uid] = listener
	return uid
}

//Remove - Remove a connection listener
func (conn *WebSocketConnection) Remove(uid string) error {
	conn.listenSync.Lock()
	defer conn.listenSync.Unlock()
	if _, ok := conn.listeners[uid]; ok {
		delete(conn.listeners, uid)
		return nil
	}
	return fmt.Errorf("%s listener not exist", uid)
}

//Send - write message to connection
func (conn *WebSocketConnection) Send(msg string) error {
	if conn.closed {
		return fmt.Errorf("connection closed")
	}
	return conn.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

//Close - closes connection
func (conn *WebSocketConnection) Close() error {
	return conn.conn.Close()
}

func (conn *WebSocketConnection) String() string {
	if conn.uuid == uuid.Nil {
		conn.uuid = uuid.New()
	}
	return conn.uuid.String()
}

func (conn *WebSocketConnection) recv() {
	for len(conn.listeners) < 1 {
	}
	for {
		mt, msg, err := conn.conn.ReadMessage()
		if err != nil {
			conn.closed = true
			for _, v := range conn.listeners {
				go v(conn, "close", "")
			}
			break
		}
		if mt != websocket.TextMessage {
			continue
		}
		for _, v := range conn.listeners {
			go v(conn, "message", string(msg))
		}
	}
}

//WebSocketConnectionFactory - a web socket connection factory
type WebSocketConnectionFactory struct {
	newListener func(Connection, map[string]string)
}

//New - add a listener for new connection
func (factory *WebSocketConnectionFactory) New(listener func(Connection, map[string]string)) {
	factory.newListener = listener
}

//Setup - setup ConnectionFactory at Get endpoint
func (factory *WebSocketConnectionFactory) Setup(getParams func(*websocket.Conn) map[string]string) func(*fiber.Ctx) {
	return websocket.New(func(c *websocket.Conn) {
		connection := &WebSocketConnection{conn: c}
		params := getParams(c)
		go factory.newListener(connection, params)
		connection.recv()
	})
}
