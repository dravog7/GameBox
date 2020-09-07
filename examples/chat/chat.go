package main

import (
	"encoding/json"
	"fmt"

	"github.com/dravog7/GameBox/connection"
)

//ChatRoom - A basic chat room
type ChatRoom struct {
	Name        string
	connections map[string]connection.Connection
	uid2name    map[string]string
	name2uid    map[string]string
}

//ChatRoomRequest - Request JSON format
type ChatRoomRequest struct {
	Command string
	Value   string
}

//ChatRoomResponse - Response JSON format
type ChatRoomResponse struct {
	Command string
	User    string
	Value   string
	Ping    string
}

//Join - join a connection to room
func (room *ChatRoom) Join(conn connection.Connection) {
	if room.connections == nil {
		room.connections = make(map[string]connection.Connection)
	}
	params := conn.GetParams()
	if ouid, ok := room.name2uid[params["name"]]; (ok) && (params["name"] != "") {
		if room.connections[ouid].IsClosed() {
			room.connections[ouid] = room.connections[ouid].Reconnect(conn)
			return
		}
		conn.SendJSON(&ChatRoomResponse{
			Command: "message",
			User:    conn.String(),
			Value:   params["name"] + " already logged in-Logged as anonymous",
		})
	}
	room.sendAllJSON(&ChatRoomResponse{
		Command: "message",
		User:    conn.String(),
		Value:   " Joined!",
	})
	room.connections[conn.String()] = conn
	conn.Listen(func(co connection.Connection, mt string, msg string) {
		req := &ChatRoomRequest{}
		if mt == "disconnect" {
			req.Command = "disconnect"
		} else if mt == "reconnect" {
			req.Command = "reconnect"
		} else {
			json.Unmarshal([]byte(msg), &req)
		}
		room.process(co, req)
	})
}

func (room ChatRoom) String() string {
	return room.Name
}

func (room *ChatRoom) getname(uid string) string {
	if room.uid2name == nil {
		room.uid2name = make(map[string]string)
		room.name2uid = make(map[string]string)
	}
	if name, ok := room.uid2name[uid]; ok {
		return name
	}
	return uid
}

func (room *ChatRoom) setname(uid string, name string) {
	if room.uid2name == nil {
		room.uid2name = make(map[string]string)
		room.name2uid = make(map[string]string)
	}
	if _, ok := room.name2uid[name]; ok {
		room.connections[uid].SendJSON(&ChatRoomResponse{
			Command: "error",
			User:    room.getname(uid),
			Value:   name + " exists",
		})
	}
	room.uid2name[uid] = name
	room.name2uid[name] = uid
	room.connections[uid].SendJSON(&ChatRoomResponse{
		Command: "remember",
		User:    room.getname(uid),
		Value:   name,
	})
}
func (room *ChatRoom) isReconnect(name string) {

}
func (room *ChatRoom) process(co connection.Connection, req *ChatRoomRequest) {
	resp := &ChatRoomResponse{
		User: room.getname(co.String()),
		Ping: co.GetPing().String(),
	}
	switch req.Command {
	case "rename":
		room.setname(co.String(), req.Value)
		resp.Command = "message"
		resp.Value = "renamed to:" + req.Value
	case "disconnect":
		resp.Command = "message"
		resp.Value = "disconnect"
	case "reconnect":
		resp.Command = "message"
		resp.Value = "reconnect"
	case "message":
		resp.Command = "message"
		resp.Value = req.Value
	}
	room.sendAllJSON(resp)
}
func (room *ChatRoom) sendAllJSON(msg interface{}) {
	for _, v := range room.connections {
		err := v.SendJSON(msg)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
