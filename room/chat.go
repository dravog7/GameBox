package room

import (
	"fmt"

	"github.com/dravog7/GameBox/connection"
)

//ChatRoom - A basic chat room
type ChatRoom struct {
	Name        string
	connections map[string]connection.Connection
}

//Join - join a connection to room
func (room *ChatRoom) Join(conn connection.Connection) {
	if room.connections == nil {
		room.connections = make(map[string]connection.Connection)
	}
	room.connections[conn.String()] = conn
	conn.Listen(func(co connection.Connection, mt string, msg string) {
		if mt == "close" {
			delete(room.connections, co.String())
			return
		}
		room.process(msg)
	})
}

func (room ChatRoom) String() string {
	return room.Name
}

func (room *ChatRoom) process(msg string) {
	for _, v := range room.connections {
		err := v.Send(msg)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
