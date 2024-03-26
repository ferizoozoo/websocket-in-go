package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
)

type Server struct {
	ip   string
	port int
}

func (s *Server) WithIp(ip string) *Server {
	s.ip = ip
	return s
}

func (s *Server) WithPort(port int) *Server {
	s.port = port
	return s
}

func New(ip string, port int) *Server {
	return &Server{ip: ip, port: port}
}

func (s *Server) Run() {
	addr := fmt.Sprintf("%s:%d", s.ip, s.port)
	server, err := net.Listen("tcp", addr)

	if err != nil {
		fmt.Println("Server could not be opened")
		os.Exit(0)
	}

	fmt.Printf("Server listening on %s:%d\n", s.ip, s.port)

	defer server.Close()

	// handle client requests
	for {
		connection, err := server.Accept()

		if err != nil {
			fmt.Println("Connection failed")
			continue
		}

		go handleRequest(connection)
	}
}

// TODO: first this function should be refactored
//
//	second, it's not complete yet
func handleRequest(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 15)

	_, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return
	}
	isGet, _ := regexp.MatchString("^GET", string(buf[:]))

	if isGet {
		fmt.Println("GET request")
		newBuf := make([]byte, 256)
		copy(newBuf[0:20], buf[:])
		buf = newBuf

		for {
			_, errRead := conn.Read(buf)
			if errRead != nil {
				break
			}

			headers := getHeaders(buf)
			SecWebSocketAccept := generateSecWebSocketAccept(headers["Sec-WebSocket-Key"])

			response := []byte("HTTP/1.1 101 Switching Protocols\r\n" +
				"Upgrade: websocket\r\n" +
				"Connection: Upgrade\r\n" +
				"Sec-WebSocket-Accept: " + SecWebSocketAccept + "\r\n" +
				"User-Agent: Anything\r\n\r\n")

			n, _ := conn.Write(response)
			fmt.Printf("Sent %d bytes\n", n)
		}

		return
	}

	// TODO: do the message receiving phase (frames, decoding, ...) after handshake

}

func generateSecWebSocketAccept(key string) string {
	hasher := sha1.New()
	hasher.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func getHeaders(buf []byte) map[string]string {
	headers := make(map[string]string)
	for _, headerLine := range strings.Split(string(buf), "\r\n") {
		if kv := strings.SplitN(headerLine, ":", 2); len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}
