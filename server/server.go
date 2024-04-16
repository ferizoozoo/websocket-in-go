package websocket

import (
	"fmt"
	"io"
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

	buf := make([]byte, 15)

	_, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return
	}

	newBuf := make([]byte, 256)
	copy(newBuf[0:20], buf[:])
	buf = newBuf

	isGet, _ := regexp.MatchString("^GET", string(buf[:]))

	if isGet {
		fmt.Println("GET request")

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
	buf = make([]byte, 1)

	_, err = conn.Read(buf)
	if err != nil {
		return
	}

	firstByte := buf[0]
	//fin := firstByte >> 7
	//rsv1 := firstByte >> 6
	//rsv2 := firstByte >> 5
	//rsv3 := firstByte >> 4
	opcode := firstByte & 7
	getInstructionFromOpcode(opcode)

	// TODO: read payload and do further process
	_, err = conn.Read(buf)
	if err != nil {
		return
	}

	payloadLength := uint64((buf[0] << 1) >> 1)

	if payloadLength == 126 {
		newBuf := make([]byte, 2)
		copy(newBuf[0:2], buf[:])
		buf = newBuf

		_, err = conn.Read(buf)
		if err != nil {
			return
		}

		payloadLength = uint64(buf[0])
		payloadLength <<= 8
		payloadLength |= uint64(buf[1])
	} else if payloadLength == 127 {
		newBuf := make([]byte, 8)
		copy(newBuf[0:2], buf[:])
		buf = newBuf

		_, err = conn.Read(buf)
		if err != nil {
			return
		}

		payloadLength = uint64(buf[0])
		payloadLength <<= 8
		payloadLength |= uint64(buf[1])
		payloadLength <<= 8
		payloadLength |= uint64(buf[2])
		payloadLength <<= 8
		payloadLength |= uint64(buf[3])
		payloadLength <<= 8
		payloadLength |= uint64(buf[4])
		payloadLength <<= 8
		payloadLength |= uint64(buf[5])
		payloadLength <<= 8
		payloadLength |= uint64(buf[6])
		payloadLength <<= 8
		payloadLength |= uint64(buf[7])
	}
}

func getInstructionFromOpcode(opcode byte) {
	switch opcode {
	case 0x0:
		// TODO: continuation
	case 0x1:
		// TODO: text
	case 0x2:
		// TODO: binary
	}
}

// TODO: decode bytes for getting payload length (refactoring the horrible code in handleRequest)
func getPayloadLengthFromConnection(conn net.Conn) int {
	panic("not implemented yet")
}
