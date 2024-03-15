package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"socket-chat/pkg/utils"
	"syscall"
)

var nickname string

func readInput(udpFd int, tcpFd int, multicastFd int) {
	reader := bufio.NewReader(os.Stdin)
	for {
		message, err := reader.ReadString('\n')

		if err != nil {
			continue
		}
		message = message[:len(message)-1]

		if len(message) == 0 {
			continue
		}

		// if message first character is "U" then send it using UDP
		if message[0] == 'U' {
			sendUsingUDP(udpFd, message)
		} else if message[0] == 'M' {
			sendUsingMulticast(multicastFd, message)
		} else if message == "E" {
			exitChat(tcpFd, udpFd, multicastFd)
		} else if message == "/ASCII" {
			sendASCIIArt(udpFd, nickname)
		} else {
			message = nickname + ": " + message

			_, err := syscall.Write(tcpFd, []byte(message+"\n"))
			utils.HandleError("Error writing message to TCP", err)
		}
	}
}

func sendASCIIArt(udpFd int, nickname string) {
	file, err := os.Open("ascii.txt")
	utils.HandleError("Error opening file", err)

	scanner := bufio.NewScanner(file)
	asciiArt := ""
	for scanner.Scan() {
		asciiArt += scanner.Text()
		message := nickname + " using UDP:" + scanner.Text()

		// send it using UDP
		addr := syscall.SockaddrInet4{Port: utils.SERVER_PORT}
		copy(addr.Addr[:], net.ParseIP(utils.SERVER_ADDR).To4())

		err = syscall.Sendto(udpFd, []byte(message), 0, &addr)
		utils.HandleError("Error sending message", err)
	}

}

func exitChat(tcpFd int, udpFd int, multicastFd int) {
	fmt.Println("Exiting chat")

	// send message about leaving
	_, err := syscall.Write(tcpFd, []byte(nickname+" has left the chat\n"))
	utils.HandleError("Error writing left message to TCP", err)

	// send message about leaving udp
	err = syscall.Sendto(udpFd, []byte("client: close"), 0, &syscall.SockaddrInet4{Port: utils.SERVER_PORT})
	utils.HandleError("Error writing left message to UDP", err)

	// Close the sockets
	syscall.Close(tcpFd)
	syscall.Close(udpFd)
	syscall.Close(multicastFd)
	os.Exit(0)
}

func sendUsingMulticast(multicastFd int, message string) {
	//send using multicast
	fmt.Print("Sending message using multicast\n")
	message = message[1:]
	message = nickname + ": " + message

	addr := syscall.SockaddrInet4{
		Port: utils.MULTICAST_PORT,
		Addr: utils.ParseIP(utils.MULTICAST_GROUP),
	}

	err := syscall.Sendto(multicastFd, []byte(message), 0, &addr)
	utils.HandleError("Error sending message", err)
}

func sendUsingUDP(udpFd int, message string) {
	fmt.Print("Sending message using UDP\n")

	// remove the first character
	message = message[1:]

	message = nickname + ": " + message
	// send it using UDP
	addr := syscall.SockaddrInet4{Port: utils.SERVER_PORT}
	copy(addr.Addr[:], net.ParseIP(utils.SERVER_ADDR).To4())

	err := syscall.Sendto(udpFd, []byte(message), 0, &addr)
	utils.HandleError("Error sending message", err)
}

func readUDP(fd int) {
	buffer := make([]byte, 1024) // Adjust the buffer size as needed
	for {
		n, _, err := syscall.Recvfrom(fd, buffer, 0)

		if err != nil {
			continue
		}

		message := string(buffer[:n])
		fmt.Println(message)
	}
}

func readMulticast(fd int) {
	buffer := make([]byte, 1024) // Adjust the buffer size as needed
	for {
		n, _, err := syscall.Recvfrom(fd, buffer, 0)

		if err != nil {
			continue
		}

		//if the message is from the same client, ignore it
		if string(buffer[:len(nickname)]) == nickname {
			continue
		}

		message := string(buffer[:n])
		fmt.Println(message)
	}
}

func readTcp(tcpFd int, udpFd int, multicastFd int, reader *bufio.Reader) {
	for {
		message, err := reader.ReadString('\n')

		if message == "server: close\n" {
			fmt.Println("Server disconnected.")

			// Close the sockets
			syscall.Close(tcpFd)
			syscall.Close(udpFd)
			syscall.Close(multicastFd)

			os.Exit(0)
		}

		if err != nil {
			fmt.Println("Client closing...")
			// Close the sockets
			syscall.Close(tcpFd)
			syscall.Close(udpFd)
			syscall.Close(multicastFd)

			os.Exit(0)
		} else {
			fmt.Print(message)
		}
	}
}

func getNickname(reader *bufio.Reader) string {
	message, err := reader.ReadString('\n')
	utils.HandleError("Error reading", err)

	nickname = message[:len(message)-1]
	fmt.Println("Your nickname is:", nickname)
	return nickname
}

func createTcpSocket() int {
	tcpFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	utils.HandleError("Error defining socket", err)

	addr := syscall.SockaddrInet4{Port: utils.SERVER_PORT}
	copy(addr.Addr[:], net.ParseIP(utils.SERVER_ADDR).To4())

	err = syscall.Connect(tcpFd, &addr)
	utils.HandleError("Error connecting to TCP", err)

	return tcpFd
}

func createUdpSocket() (int, syscall.SockaddrInet4) {
	udpFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	utils.HandleError("Error defining socket", err)

	addr := syscall.SockaddrInet4{Port: utils.SERVER_PORT}
	copy(addr.Addr[:], net.ParseIP(utils.SERVER_ADDR).To4())

	return udpFd, addr
}

func connectToMulticastGroup() int {
	multicastFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	utils.HandleError("Error defining socket", err)

	err = syscall.SetsockoptInt(multicastFd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	utils.HandleError("Error adding socket property", err)

	err = syscall.SetsockoptInt(multicastFd, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	utils.HandleError("Error adding socket property", err)

	mreq := syscall.IPMreq{
		Multiaddr: utils.ParseIP(utils.MULTICAST_GROUP),
		Interface: utils.ParseIP(utils.INTERFACE),
	}

	err = syscall.SetsockoptIPMreq(multicastFd, syscall.IPPROTO_IP, syscall.IP_ADD_MEMBERSHIP, &mreq)
	utils.HandleError("Error adding multicast group membership", err)

	addr := syscall.SockaddrInet4{
		Port: utils.MULTICAST_PORT,
		Addr: utils.ParseIP(utils.INTERFACE),
	}

	err = syscall.Bind(multicastFd, &addr)
	utils.HandleError("Error binding socket", err)

	return multicastFd
}

func main() {
	fmt.Println("Starting client...")

	// Create TCP socket and connect to server
	tcpFd := createTcpSocket()

	//get first message from server and set it as nickname
	reader := bufio.NewReader(os.NewFile(uintptr(tcpFd), "server"))
	nickname = getNickname(reader)

	// Create UDP socket and send join message to server
	udpFd, addr := createUdpSocket()

	err := syscall.Sendto(udpFd, []byte(nickname+" has joined the chat"), 0, &addr)
	utils.HandleError("Error sending join message", err)

	// Connect to multicast group
	multicastFd := connectToMulticastGroup()

	// Start reading input and messages
	go readInput(udpFd, tcpFd, multicastFd)

	go readUDP(udpFd)

	go readMulticast(multicastFd)

	go readTcp(tcpFd, udpFd, multicastFd, reader)

	// Wait for interrupt signal to exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	// send message about leaving to unicast and tcp
	_, err = syscall.Write(tcpFd, []byte("client: close\n"))
	utils.HandleError("Error writing left message to TCP", err)

	err = syscall.Sendto(udpFd, []byte("client: close"), 0, &syscall.SockaddrInet4{Port: utils.SERVER_PORT})
	utils.HandleError("Error writing left message to UDP", err)

	// Close the sockets
	syscall.Close(udpFd)
	syscall.Close(multicastFd)
	syscall.Close(tcpFd)
	os.Exit(0)
}
