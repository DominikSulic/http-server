package httpserver

import (
	"fmt"
	"io"
	"log"
	"net"

	"http-server/internal/request"
	"http-server/internal/response"
)

type HandlerError struct {
	Message    string
	StatusCode response.StatusCode
}

type ResponseHandler func(writer *response.Writer, request *request.HttpRequest)

type Server struct {
	closed          bool
	responseHandler ResponseHandler
}

func Start(port uint16, responseHandler ResponseHandler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		closed:          false,
		responseHandler: responseHandler,
	}

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

	responseWriter := response.NewWriter(connection)
	httpRequest, err := request.RequestFromReader(connection)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(*response.GetDefaultHeaders(0))
		return
	}

	server.responseHandler(responseWriter, httpRequest)
}
