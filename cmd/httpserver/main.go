package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"http-server/internal/httpserver"
	"http-server/internal/request"
	"http-server/internal/response"
)

const port = 42069

func main() {
	server, err := httpserver.Start(port, responseHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	defer server.Close()
	log.Println("Server started on port", port)

	// common pattern in Go for gracefully shutting down a server. Because server.Server returns immediately (it handles requests in the background in goroutines),
	// if you exit main immediately, the server will just stop. You want to wait for a signal (i.e. CTRL + C) before you stop the server.
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	<-signalChannel
	log.Println("Server gracefully stopped")
}

func responseHandler(writer io.Writer, request *request.HttpRequest) *httpserver.HandlerError {
	if request.RequestLine.RequestTarget == "/badrequest" {
		return &httpserver.HandlerError{StatusCode: response.StatusBadRequest, Message: "Bad request\n"}
	}

	if request.RequestLine.RequestTarget == "/internalerror" {
		return &httpserver.HandlerError{StatusCode: response.StatusInternalServerError, Message: "Internal error\n"}
	}

	writer.Write([]byte("All good\n"))
	return nil
}
