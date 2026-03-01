package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddress, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("Error resolving udp address: ", err)
	}

	// create a udp connection, though it is essentially conectionless
	connection, err := net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		log.Fatal("error while creating udp connection using Dial", err)
	}

	defer connection.Close()

	// creates a new reader that reads from standard input
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(">")

		// \n is the delimiter
		userInput, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("error while reading user input: ", err)
		}

		_, err = connection.Write([]byte(userInput))
		if err != nil {
			log.Printf("error writing user input to connection: ", err)
			return
		}

	}
}
