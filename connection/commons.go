package connection

import "log"

func safeConnEmit(v func(Connection, string, string), conn Connection, info string, msg string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered emit(WebSocketConnection)", r)
		}
	}()
	v(conn, info, msg)
}
