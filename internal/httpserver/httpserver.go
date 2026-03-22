package httpserver

import (
	"fmt"
	"io"
	"log"
	"net"

	"http-server/internal/response"
)

type Server struct {
	closed bool
}

func Start(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{}
	server.closed = false

	go server.listen(listener)

	return server, nil
}

func (server *Server) Close() error {
	server.closed = true
	return nil // TODO: ? this is weird
}

func (server *Server) listen(listener net.Listener) {
	for {
		connection, err := listener.Accept()

		if server.closed {
			return
		}

		if err != nil {
			log.Fatal("Error accepting connection:\n", err)
		}

		go server.handle(connection)
	}
}

func (server *Server) handle(connection io.ReadWriteCloser) {
	defer connection.Close()

	headers := response.GetDefaultHeaders(0)

	// output := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello world!")

	response.WriteStatusLine(connection, response.StatusOK)

	response.WriteHeaders(connection, headers)

	connection.Close()
}
