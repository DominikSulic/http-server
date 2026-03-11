package main

import (
	"fmt"
	"log"
	"net"

	"http-server/internal/request"
)

/*
 * nc -u -l 42069
 * go run ./cmd/tcplistener | tee rawget.http
 * curl http://localhost:42069/pizza
 * curl -X POST http://localhost:42069/pizza -H "Content-Type: application/json" -d '{"flavor":"dark mode"}'
 * go run ./cmd/tcplistener | tee /tmp/requestline.txt
 */
func main() {
	// listen on TCP port 42069 on all IP addresses of the local system
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("Error creating the listener:\n", err)
	}

	for {

		connection, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting connection:\n", err)
		}

		httpRequest, err := request.RequestFromReader(connection)
		if err != nil {
			log.Fatal("Error requesting from reader: ", err)
		}

		fmt.Println("Request Line:")
		fmt.Printf("-Method: %s\n", httpRequest.RequestLine.Method)
		fmt.Printf("-Target: %s\n", httpRequest.RequestLine.RequestTarget)
		fmt.Printf("-Version: %s\n", httpRequest.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")

		httpRequest.Headers.ForEach(func(key, value string) {
			fmt.Printf("- %s: %s\n", key, value)
		})
	}
}
