package httpserver

import (
	"bytes"
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

type ResponseHandler func(writer io.Writer, request *request.HttpRequest) *HandlerError

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
	headers := response.GetDefaultHeaders(0)
	httpRequest, err := request.RequestFromReader(connection)
	if err != nil {
		response.WriteStatusLine(connection, response.StatusBadRequest)
		response.WriteHeaders(connection, headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerError := server.responseHandler(writer, httpRequest)

	var body []byte = nil
	var status response.StatusCode = response.StatusOK
	if handlerError != nil {
		status = handlerError.StatusCode
		body = []byte(handlerError.Message)
	} else {
		body = writer.Bytes()
	}

	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	response.WriteStatusLine(connection, status)
	response.WriteHeaders(connection, headers)
	connection.Write(body)
}
