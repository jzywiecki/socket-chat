package utils

import (
	"fmt"
	"math/rand"
	"net"
)

const (
	SERVER_ADDR     = "127.0.0.1"
	MAX_MSG_SIZE    = 1024
	SERVER_PORT     = 8084
	MULTICAST_GROUP = "224.0.0.4"
	MULTICAST_PORT  = 8082
	MULTICAST_TTL   = 5
	MAX_CONNECTIONS = 30
	INTERFACE       = "0.0.0.0"
)

func HandleError(message string, err error) {
	if err != nil {
		fmt.Println(message, err)
		panic(err)
	}
}

func RandomNumber(len int) int {
	return rand.Intn(len)
}

func ParseIP(ip string) [4]byte {
	var addr [4]byte
	copy(addr[:], net.ParseIP(ip).To4())
	return addr
}
