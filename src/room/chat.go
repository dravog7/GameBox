package room

import (
	"gamebox/connection"
)

//ChatRoom - A basic chat room
type ChatRoom struct {
	Name        string
	msgs        []string
	connections []connection.Connection
}

//Join - join a connection to room
func (room *ChatRoom) Join(conn connection.Connection) {
	room.connections = append(room.connections, conn)
	conn.Listen(func(co connection.Connection, mt string, msg string) {
		if mt == "close" {
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
			continue
		}
	}
	room.msgs = append(room.msgs, msg)
}
