package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"http-server/internal/httpserver"
	"http-server/internal/request"
	"http-server/internal/response"
)

const port = 42069

func respond200() []byte {
	return []byte(`<html>
				   	   <head>
					       <title>200 OK</title>
					   </head>
					   <body>
					       <h1>Success!</h1>
						   <p>The request provided was good.</p>
					   </body>
				   </html>`)
}

func respond400() []byte {
	return []byte(`<html>
				   	   <head>
					       <title>400 Bad Request</title>
					   </head>
					   <body>
					       <h1>Bad Request</h1>
						   <p>The request provided was bad.</p>
					   </body>
				   </html>`)
}

func respond500() []byte {
	return []byte(`<html>
				   	   <head>
					       <title>500 Internal Server Error</title>
					   </head>
					   <body>
					       <h1>Internal Server Error</h1>
						   <p>This one's my bad.</p>
					   </body>
				   </html>`)
}

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

func responseHandler(writer *response.Writer, request *request.HttpRequest) {
	status := response.StatusOK
	body := respond200()

	if request.RequestLine.RequestTarget == "/badrequest" {
		status = response.StatusBadRequest
		body = respond400()
	}

	if request.RequestLine.RequestTarget == "/internalerror" {
		status = response.StatusInternalServerError
		body = respond500()
	}

	headers := response.GetDefaultHeaders(0)
	headers.Replace("Content-Type", "text/html")
	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	writer.WriteStatusLine(status)
	writer.WriteHeaders(*headers)
	writer.WriteBody(body)
}
