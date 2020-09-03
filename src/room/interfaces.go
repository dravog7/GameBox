package room

import (
	"gamebox/connection"
)

//Room - Defines a room where the core logic of game resides
type Room interface {
	Join(connection.Connection)
	String() string //room name as string
}

//Manager - Manages supply of Connections to rooms
type Manager interface {
	Register(Room) error
	AddFactory(connection.Factory, func(error))
}
