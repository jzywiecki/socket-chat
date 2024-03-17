# Simple sockets chat
Simple chat using UDP and TCP sockets, being able to send messages to multicast group to avoid communication with server. 


In order to run server and clients you should use LINUX/macOS and Go 1.22.

Client:\
`go run ./pkg/client/client.go`

Server:\
`go run ./pkg/server/server.go`

Joining and sending messages:\
<img width="230" alt="image" src="https://github.com/jzywiecki/socket-chat/assets/105950890/8268fc92-0da8-4f7d-b204-7d7471581e0e">

Recieving messages:\
<img width="455" alt="image" src="https://github.com/jzywiecki/socket-chat/assets/105950890/48a602c5-d2b6-4953-8dd2-6425a535f679">

Server:\
<img width="716" alt="image" src="https://github.com/jzywiecki/socket-chat/assets/105950890/5b74d0b1-7e33-4efc-bb22-092cd0d8284e">


Commands:\
`M<message>` - multicast\
`U<message>` - unicast\
`/ASCII` - ascii art\
`E` - exiting client, you can also use SIGINT\
