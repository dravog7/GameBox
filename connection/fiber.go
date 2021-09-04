package connection

import (
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type FiberWebSocketConnection struct {
	conn *websocket.Conn
	uuid uuid.UUID

	closed bool

	*Subscription

	ping     time.Duration
	interval time.Duration
}

// pingLoop runs infinitely to check ping every 10 secs
func (conn *FiberWebSocketConnection) pingLoop() {
	conn.conn.SetPongHandler(func(dateText string) error {
		var then time.Time
		then.UnmarshalText([]byte(dateText))
		conn.ping = time.Since(then)
		return nil
	})
	for !conn.closed {
		data, _ := time.Now().MarshalText()
		err := conn.conn.WriteControl(websocket.PingMessage, data, time.Now().Add(time.Second*10))
		if websocket.IsCloseError(err) {
			break
		}
		time.Sleep(conn.interval)
	}
}

func (conn *FiberWebSocketConnection) GetPing() time.Duration {
	return conn.ping
}

func (conn *FiberWebSocketConnection) Send(msg string) error {
	return conn.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (conn *FiberWebSocketConnection) SendJSON(msg interface{}) error {
	return conn.conn.WriteJSON(msg)
}

func (conn *FiberWebSocketConnection) recv() {

	// wait for subscribers to exist
	for len(conn.subscribers) < 1 {
	}

	for {
		mt, msg, err := conn.conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			conn.emit(CloseMessage, err.Error())
			conn.closed = true
			conn.closeSubscription()
			break
		} else if err != nil {
			conn.emit(ErrorMessage, err.Error())
			continue
		}
		if mt != websocket.TextMessage {
			continue
		}

		conn.emit(TextMessage, string(msg))
	}
}

func (conn *FiberWebSocketConnection) IsClosed() bool {
	return conn.closed
}

func (conn *FiberWebSocketConnection) Close() error {
	return conn.conn.Close()
}

func (conn *FiberWebSocketConnection) GetUUID() uuid.UUID {
	return conn.uuid
}

func NewFiberWebSocketConnection(conn *websocket.Conn) *FiberWebSocketConnection {
	return &FiberWebSocketConnection{
		conn,
		uuid.New(),
		false,
		&Subscription{make(map[uuid.UUID]chan *InMessage)},
		time.Duration(0),
		time.Second * 10,
	}
}

func (conn *FiberWebSocketConnection) StartListeners() {
	go conn.pingLoop()
	//exiting will close the connection
	conn.recv()
}
