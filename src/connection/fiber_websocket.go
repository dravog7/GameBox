package connection

import (
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/gofiber/websocket"
	"github.com/google/uuid"
)

//WebSocketConnection - a web socket connection struct
type WebSocketConnection struct {
	conn      *websocket.Conn
	closed    bool
	listeners []func(Connection, string, string)
	uuid      uuid.UUID
}

//Listen - add listener for messages
func (conn *WebSocketConnection) Listen(listener func(Connection, string, string)) {
	conn.listeners = append(conn.listeners, listener)
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
func (factory *WebSocketConnectionFactory) Setup() func(*fiber.Ctx) {
	return websocket.New(func(c *websocket.Conn) {
		connection := &WebSocketConnection{conn: c}
		params := map[string]string{
			"id": c.Params("id"),
		}
		go factory.newListener(connection, params)
		connection.recv()
	})
}
