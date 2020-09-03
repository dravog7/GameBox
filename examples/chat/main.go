package main

import (
	"fmt"

	"github.com/dravog7/GameBox/connection"
	"github.com/dravog7/GameBox/room"

	"github.com/gofiber/fiber"
	"github.com/gofiber/websocket"
)

func main() {
	app := fiber.New()

	app.Static("/", "./statics")
	app.Use(func(c *fiber.Ctx) {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			c.Next()
		}
	})
	chatroom := &ChatRoom{Name: "1"}
	manager := room.DefaultManager{}
	manager.Register(chatroom)
	factory := &connection.WebSocketConnectionFactory{}
	manager.AddFactory(factory, func(err error) {
		fmt.Println(err)
	})
	app.Get("/ws/:id", factory.Setup())
	app.Listen(3000)
}
