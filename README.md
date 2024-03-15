Aby uruchomić serwer i klientów nalezy korzystając z systemu LINUX/macOS (na tych systemach rozwiązanie jest dostosowane względem systemowych struktur zwracanych przez operacje na socketach), oraz języka Go (Golang) w wersji 1.22 uruchomić klienta i serwer.

Klient:
`go run ./pkg/client/client.go`

Serwer:
`go run ./pkg/server/server.go`

Dostępne komendy to:
`M<wiadomosc>` - multicast
`U<wiadomosc>` - unicast
`/ASCII` - wysłanie ascii art
`E` - wyjscie z klienta, programy są przygotowane do uzywania SIGTERM