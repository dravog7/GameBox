package room

import (
	"fmt"
	"gamebox/connection"
)

//DefaultManager - The default room manager
type DefaultManager struct {
	rooms map[string]Room
}

//Register - register a new room
func (manager *DefaultManager) Register(room Room) error {
	if manager.rooms == nil {
		manager.rooms = make(map[string]Room)
	}
	_, ok := manager.rooms[room.String()]
	if ok {
		return fmt.Errorf("room already exists")
	}
	manager.rooms[room.String()] = room
	return nil
}

//AddFactory - adds a connection factory for manager to listen to
func (manager DefaultManager) AddFactory(factory connection.Factory, callback func(error)) {
	factory.New(func(conn connection.Connection, params map[string]string) {
		room, ok := manager.rooms[params["id"]]
		if ok {
			room.Join(conn)
		} else {
			go callback(fmt.Errorf("room %v not found", params["id"]))
		}
	})
}
