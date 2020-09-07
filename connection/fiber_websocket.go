package connection

import (
	"fmt"
	"sync"
	"time"

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
	writeSync  sync.Mutex
	closed     bool
	listeners  map[string]func(Connection, string, string)
	listenSync sync.Mutex
	uuid       string
	params     map[string]string
	ping       time.Duration
}

//Listen - add listener for messages
func (conn *WebSocketConnection) Listen(listener func(Connection, string, string)) string {
	conn.listenSync.Lock()
	defer conn.listenSync.Unlock()
	uid := uuid.New().String()
	// for _, ok := conn.listeners[uid]; ok; {
	// 	uid = uuid.New().String()
	// 	_, ok = conn.listeners[uid]
	// }
	conn.listenTo(listener, uid)
	return uid
}
func (conn *WebSocketConnection) listenTo(listener func(Connection, string, string), uid string) {
	if conn.listeners == nil {
		conn.listeners = make(map[string]func(Connection, string, string))
	}
	conn.listeners[uid] = listener
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
	conn.writeSync.Lock()
	defer conn.writeSync.Unlock()
	if conn.closed {
		return fmt.Errorf("connection closed")
	}
	return conn.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

//SendJSON - write JSON message to connection
func (conn *WebSocketConnection) SendJSON(msg interface{}) error {
	conn.writeSync.Lock()
	defer conn.writeSync.Unlock()
	if conn.closed {
		return fmt.Errorf("connection closed")
	}
	return conn.conn.WriteJSON(msg)
}

//Close - closes connection
func (conn *WebSocketConnection) Close() error {
	return conn.conn.Close()
}

//IsClosed - check if connection is closed
func (conn *WebSocketConnection) IsClosed() bool {
	return conn.closed
}

//GetParams - returns params map from setup()
func (conn *WebSocketConnection) GetParams() map[string]string {
	return conn.params
}
func (conn *WebSocketConnection) String() string {
	if conn.uuid == "" {
		conn.uuid = uuid.New().String()
	}
	return conn.uuid
}

func (conn *WebSocketConnection) recv() {
	go conn.pingLoop()
	for len(conn.listeners) < 1 {
	}
	for {
		mt, msg, err := conn.conn.ReadMessage()
		if err != nil {
			conn.closed = true
			conn.emit("disconnect", "")
			break
		}
		if mt != websocket.TextMessage {
			continue
		}
		conn.emit("message", string(msg))
	}
}

func (conn *WebSocketConnection) pingLoop() {
	conn.conn.SetPongHandler(func(dateText string) error {
		var then time.Time
		then.UnmarshalText([]byte(dateText))
		conn.ping = time.Now().Sub(then)
		return nil
	})
	for !conn.closed {
		data, _ := time.Now().MarshalText()
		conn.conn.WriteControl(websocket.PingMessage, data, time.Now().Add(time.Second*10))
		time.Sleep(time.Second * 10)
	}
}

func (conn *WebSocketConnection) emit(info string, msg string) {
	for _, v := range conn.listeners {
		go safeConnEmit(v, conn, info, msg)
	}
}

//Reconnect - Copy listeners of self to Connection c [not delete]
func (conn *WebSocketConnection) Reconnect(c Connection) Connection {
	//accept connection interface type, assign all listeners to it
	conn.listenSync.Lock()
	defer conn.listenSync.Unlock()
	conn.emit("reconnect", "")
	for uid, listener := range conn.listeners {
		c.listenTo(listener, uid)
	}
	c.setUID(conn.uuid)
	return c
}
func (conn *WebSocketConnection) setUID(uid string) {
	conn.uuid = uid
}

//GetPing - return the ping of connection
func (conn *WebSocketConnection) GetPing() time.Duration {
	return conn.ping
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
		var connection *WebSocketConnection
		params := getParams(c)
		connection = &WebSocketConnection{conn: c, params: params}
		go factory.newListener(connection, params)
		connection.recv() //returning closes the websocket
	})
}
