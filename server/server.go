package server

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/ferizoozoo/websocket-in-go/internal/shared"
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

func handleRequest(conn net.Conn) {
	defer conn.Close()

	buf, err := shared.ReadFromConnectionToBuffer(conn, 256)

	if err != nil {
		return
	}

	isGet, _ := regexp.MatchString("^GET", string(buf[:]))

	if isGet {
		for {
			_, errRead := conn.Read(buf)
			if errRead != nil {
				break
			}

			headers := shared.GetHeaders(buf)
			SecWebSocketAccept := shared.GenerateSecWebSocketAccept(headers["Sec-WebSocket-Key"])

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

	isWs, _ := regexp.MatchString("^ws", string(buf[:]))
	if !isWs {
		return
	}

	/* TODO: parse the frame with the following format
			First byte:
	bit 0: FIN
	bit 1: RSV1
	bit 2: RSV2
	bit 3: RSV3
	bits 4-7 OPCODE
	Bytes 2-10: payload length (see Decoding Payload Length)
	If masking is used, the next 4 bytes contain the masking key (see Reading and unmasking the data)
	All subsequent bytes are payload
	*/
	buf, _ = shared.ReadFromConnectionToBuffer(conn, 2)

	firstByte := buf[0]

	mask := firstByte >> 3

	payloadLength := uint64((buf[1] << 1) >> 1)

	if payloadLength == 126 {
		buf, _ = shared.ReadFromConnectionToBuffer(conn, 2)
		binary.BigEndian.PutUint64(buf, uint64(payloadLength))
	} else if payloadLength == 127 {
		buf, _ = shared.ReadFromConnectionToBuffer(conn, 8)
		binary.BigEndian.PutUint64(buf, uint64(payloadLength))
	}

	buf, _ = shared.ReadFromConnectionToBuffer(conn, 4)

	var maskingKey []byte

	if mask == 1 {
		buf, _ = shared.ReadFromConnectionToBuffer(conn, 4)
		maskingKey = buf
	}

	buf, _ = shared.ReadFromConnectionToBuffer(conn, int(payloadLength))

	decodedData := shared.XorEncryption(buf, maskingKey)
	conn.Write(decodedData)
}
