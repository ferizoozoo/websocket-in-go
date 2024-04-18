package main

import server "github.com/ferizoozoo/websocket-in-go/server"

func main() {
	serv := server.New("127.0.0.1", 8080)
	serv.Run()
}
