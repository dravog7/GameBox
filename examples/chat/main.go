package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dravog7/GameBox/connection"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type OutMessage struct {
	Command string
	Value   string
	Ping    string `json:",omitempty"`
}

func main() {
	app := fiber.New()
	app.Static("/", "./statics")
	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			c.Next()
			return nil
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		gameboxConnection := connection.NewFiberWebSocketConnection(c)
		subscriber := make(chan *connection.InMessage)
		gameboxConnection.Subscribe(subscriber)
		go func(subscriber chan *connection.InMessage, conn connection.IConnection) {
			for msg := range subscriber {
				var msgStruct OutMessage
				if json.Unmarshal([]byte(msg.Message), &msgStruct) != nil {
					return
				}
				fmt.Println(msgStruct)
				msgStruct.Ping = conn.GetPing().String()
				conn.SendJSON(msgStruct)
			}
		}(subscriber, gameboxConnection)
		gameboxConnection.StartListeners()
	}))
	log.Fatal(app.Listen(":3000"))
}
