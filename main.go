package main

import websocket "github.com/ferizoozoo/websocket-server-in-go/server"

func main() {
	serv := websocket.New("127.0.0.1", 8080)
	serv.Run()
}
