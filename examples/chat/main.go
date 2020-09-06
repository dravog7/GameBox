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
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			c.Next()
		}
	})

	chatroom := &ChatRoom{Name: "1"}
	manager := room.DefaultManager{}
	factory := &connection.WebSocketConnectionFactory{}

	manager.Register(chatroom)
	manager.AddFactory(factory, func(err error) {
		fmt.Println(err)
	})

	app.Get("/ws/:id", factory.Setup(func(c *websocket.Conn, ctx *connection.WebSocketParamContext) map[string]string {
		params := map[string]string{
			"id": c.Params("id"), //id used by default manager to set entry point room
		}
		flag := 0
		for {
			c.WriteMessage(websocket.TextMessage, []byte("getName"))
			_, name, err := c.ReadMessage()
			if err != nil {
				fmt.Println("error")
				break
			}
			strName := string(name)
			flag = 0
			for _, connName := range ctx.ConnectionNames {
				if connName == strName {
					flag = 1
					break
				}
			}
			if flag == 1 {
				continue
			}
			params["ConnectionName"] = strName
			break
		}

		return params
	}))
	app.Listen(3000)
}
