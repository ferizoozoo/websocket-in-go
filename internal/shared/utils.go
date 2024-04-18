package shared

import (
	"crypto/sha1"
	"encoding/base64"
	"net"
	"strings"
)

func XorEncryption(data []byte, key []byte) []byte {
	for i := 0; i < len(data); i++ {
		data[i] = data[i] ^ key[i%len(key)]
	}
	return data
}

func GenerateSecWebSocketAccept(key string) string {
	hasher := sha1.New()
	hasher.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func GetHeaders(buf []byte) map[string]string {
	headers := make(map[string]string)
	for _, headerLine := range strings.Split(string(buf), "\r\n") {
		if kv := strings.SplitN(headerLine, ":", 2); len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}

func ReadFromConnectionToBuffer(conn net.Conn, size int) ([]byte, error) {
	buf := make([]byte, size)
	_, err := conn.Read(buf)
	return buf, err
}
