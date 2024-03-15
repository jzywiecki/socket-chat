package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"socket-chat/pkg/utils"

	"github.com/go-faker/faker/v4"
)

func handleClient(connFd int, clients map[int]string) {
	defer syscall.Close(connFd)
	defer delete(clients, connFd)
	defer fmt.Println("Client", clients[connFd], "disconnected")

	reader := bufio.NewReader(os.NewFile(uintptr(connFd), "client"))
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		message = message[:len(message)-1] // Remove newline character

		if message == "client: close" {
			message = clients[connFd] + " has left the chat"
		}

		fmt.Println("Received message on TCP:", message)
		for clientFd := range clients {
			// Send message to all clients except the sender
			if clientFd != connFd {
				if _, err := syscall.Write(clientFd, []byte(message+"\n")); err != nil {
					fmt.Println("Error sending message to client:", err)
				}
			}
		}

	}
}

func handleUDP(clientsUdp []syscall.SockaddrInet4, udpFd int, c chan os.Signal) {
	for {
		select {
		case <-c:
			return
		default:
			buffer := make([]byte, utils.MAX_MSG_SIZE)
			n, senderAddr, err := syscall.Recvfrom(udpFd, buffer, 0)

			if err != nil {
				continue
			}

			sender := senderAddr.(*syscall.SockaddrInet4)
			fmt.Println("Received message on UDP:", string(buffer[:n]))

			// check if message is "client close"
			if string(buffer[:n]) == "client: close" {
				// remove client from clientsUdp
				for i, client := range clientsUdp {
					if client.Port == sender.Port && client.Addr == sender.Addr {
						clientsUdp = append(clientsUdp[:i], clientsUdp[i+1:]...)
						break
					}
				}
				continue
			}

			// Add sender to clientsUdp if he is not already in the list
			alreadyInList := false
			for _, client := range clientsUdp {
				if client.Port == sender.Port && client.Addr == sender.Addr {
					alreadyInList = true
					break
				}
			}

			if !alreadyInList {
				clientsUdp = append(clientsUdp, *sender)
			}

			// Broadcast message to all clients except the sender
			for _, client := range clientsUdp {
				if client.Port != sender.Port || client.Addr != sender.Addr {
					if err := syscall.Sendto(udpFd, buffer[:n], 0, &client); err != nil {
						fmt.Println("Error sending message to client:", err)
					}
				}
			}
		}
	}
}

func acceptConnections(sockFd int, clients map[int]string) {
	for {
		connFd, _, err := syscall.Accept(sockFd)
		if err != nil {
			continue
		}

		nickname := faker.Username()

		clients[connFd] = nickname

		// Send welcome message to client
		welcomeMessage := clients[connFd]

		if _, err := syscall.Write(connFd, []byte(welcomeMessage+"\n")); err != nil {
			fmt.Println("Error sending message to client:", err)
		}

		go handleClient(connFd, clients)
	}
}

func createTcpSocket() int {
	sockFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	utils.HandleError("Error defining socket", err)

	addr := syscall.SockaddrInet4{Port: utils.SERVER_PORT}
	copy(addr.Addr[:], net.ParseIP(utils.SERVER_ADDR).To4())

	err = syscall.Bind(sockFd, &addr)
	utils.HandleError("Error binding socket", err)

	return sockFd
}

func createUdpSocket() int {
	udpFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	utils.HandleError("Error defining UDP socket", err)

	addr := syscall.SockaddrInet4{
		Port: utils.SERVER_PORT,
		Addr: [4]byte{127, 0, 0, 1},
	}

	err = syscall.Bind(udpFd, &addr)
	utils.HandleError("Error binding UDP socket", err)

	return udpFd
}

func main() {
	fmt.Println("Starting server...")

	// Create TCP socket and bind address
	sockFd := createTcpSocket()

	// Listen for connections
	err := syscall.Listen(sockFd, utils.MAX_CONNECTIONS)
	utils.HandleError("Error listening for connections", err)

	// Map to store connected clients on TCP and list to store connected clients on UDP
	clients := make(map[int]string)
	clientsUdp := []syscall.SockaddrInet4{}

	udpFd := createUdpSocket()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// Goroutine for handling UDP messages
	go handleUDP(clientsUdp, udpFd, sigs)

	// Accept incoming connections
	go acceptConnections(sockFd, clients)

	// handle sigterm
	<-sigs

	// when sigterm we send to all clients message that server is closing letting them know to close the connection
	for clientFd := range clients {
		if _, err := syscall.Write(clientFd, []byte("server: close\n")); err != nil {
			fmt.Println("Error sending message to client:", err)
		}
	}

	// wait for clients to receive the message
	time.Sleep(500 * time.Millisecond)
	syscall.Close(sockFd) // Close our part
	syscall.Close(udpFd)  // Close our part
	os.Exit(0)
}
