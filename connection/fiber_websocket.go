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
	uuid       string
	params     map[string]string
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

//SendJSON - write JSON message to connection
func (conn *WebSocketConnection) SendJSON(msg interface{}) error {
	if conn.closed {
		return fmt.Errorf("connection closed")
	}
	return conn.conn.WriteJSON(msg)
}

//Close - closes connection
func (conn *WebSocketConnection) Close() error {
	return conn.conn.Close()
}

func (conn *WebSocketConnection) String() string {
	if conn.uuid == "" {
		conn.uuid = uuid.New().String()
	}
	return conn.uuid
}

//GetParam - Get param[key] of connection (Param was assigned by GetParam argument to setup in factory)
func (conn *WebSocketConnection) GetParam(key string) (string, bool) {
	v, err := conn.params[key]
	return v, err
}

func (conn *WebSocketConnection) recv() {
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

func (conn *WebSocketConnection) emit(info string, msg string) {
	for _, v := range conn.listeners {
		go v(conn, info, msg)
	}
}

func (conn *WebSocketConnection) reconnect(c *websocket.Conn) bool {
	if conn.closed {
		conn.conn = c
		conn.closed = false
		conn.emit("reconnect", "")
		return true
	}
	return false
}

//WebSocketConnectionFactory - a web socket connection factory
type WebSocketConnectionFactory struct {
	connections    map[string]*WebSocketConnection
	connectionSync sync.Mutex
	newListener    func(Connection, map[string]string)
}

//New - add a listener for new connection
func (factory *WebSocketConnectionFactory) New(listener func(Connection, map[string]string)) {
	factory.newListener = listener
}

//Setup - setup ConnectionFactory at Get endpoint
func (factory *WebSocketConnectionFactory) Setup(getParams func(*websocket.Conn, *WebSocketParamContext) map[string]string) func(*fiber.Ctx) {
	return websocket.New(func(c *websocket.Conn) {
		factory.connectionSync.Lock()
		if factory.connections == nil {
			factory.connections = make(map[string]*WebSocketConnection)
		}
		var connection *WebSocketConnection
		params := getParams(c, factory.newContext())
		if _, ok := params["ConnectionName"]; ok {
			if connection, ok := factory.connections[params["ConnectionName"]]; ok {
				//Reconnecting
				if connection.reconnect(c) {
					factory.connectionSync.Unlock()
					connection.params = params
					connection.recv()
				}
				return
			}
			connection = &WebSocketConnection{conn: c, uuid: params["ConnectionName"], params: params}

		} else {
			connection = &WebSocketConnection{conn: c, params: params}

		}
		factory.connections[connection.String()] = connection
		go factory.newListener(connection, params)
		factory.connectionSync.Unlock()
		connection.recv()
	})
}

//WebSocketParamContext - Context struct send to getParam param in Setup
type WebSocketParamContext struct {
	ConnectionNames []string
}

func (factory *WebSocketConnectionFactory) newContext() *WebSocketParamContext {
	var filteredNames []string
	for k, conn := range factory.connections {
		if conn.closed == false {
			filteredNames = append(filteredNames, k)
		}
	}
	return &WebSocketParamContext{
		ConnectionNames: filteredNames,
	}
}
