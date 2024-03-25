package websocket

import (
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
	bufferString := string(buf[:])
	isGet, _ := regexp.MatchString("^GET", bufferString)

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

			bufferString := string(buf[:])

			headerString := strings.Split(bufferString, "\r\n")
			headers := make(map[string]string)
			for _, headerline := range headerString {
				headerline := strings.TrimSuffix(headerline, "\n")
				headerKeyValue := strings.Split(headerline, ":")
				if len(headerKeyValue) > 1 && headerline != "" {
					key := strings.TrimSuffix(headerKeyValue[0], "\n")
					key = strings.TrimSpace(key)
					value := strings.TrimSuffix(headerKeyValue[1], "\n")
					value = strings.TrimSpace(value)
					headers[key] = value
				}
			}

			responseString := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n" +
				"Upgrade: websocket\r\n" +
				"Connection: Upgrade\r\n" +
				"Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=\r\n" +
				"User-Agent: Anything\r\n\r\n")
			response := []byte(responseString)
			n, _ := conn.Write(response)
			fmt.Printf("Sent %d bytes\n", n)
		}

		return
	}

	// TODO: do the message receiving phase (frames, decoding, ...) after handshake

}
