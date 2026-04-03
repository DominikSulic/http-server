package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"http-server/internal/headers"
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

// echo -e "GET /httpbin/stream/100 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069
func responseHandler(writer *response.Writer, request *request.HttpRequest) {
	requestTarget := request.RequestLine.RequestTarget

	if requestTarget == "/badrequest" {
		handleBadRequest(writer, request)
	} else if requestTarget == "/internalerror" {
		handleInternalError(writer, request)
	} else if strings.HasPrefix(requestTarget, "/httpbin/") {
		handleHttpBinRequestWithChunkedEncoding(writer, request)
	} else if requestTarget == "/video" {
		handleVideoRequest(writer, request)
	}
}

func handleBadRequest(writer *response.Writer, request *request.HttpRequest) {
	headers := response.GetDefaultHeaders(0)
	headers.Replace("Content-Type", "text/html")
	status := response.StatusBadRequest
	body := respond400()
	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))

	writeStatusHeadersAndBody(writer, status, headers, body)
}

func handleInternalError(writer *response.Writer, request *request.HttpRequest) {
	headers := response.GetDefaultHeaders(0)
	headers.Replace("Content-Type", "text/html")
	status := response.StatusInternalServerError
	body := respond500()
	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))

	writeStatusHeadersAndBody(writer, status, headers, body)
}

/*
run the server and debug your code with curl --raw
*/
func handleHttpBinRequestWithChunkedEncoding(writer *response.Writer, request *request.HttpRequest) {
	httpbinResponse, err := http.Get("https://httpbin.org/" + strings.TrimPrefix(request.RequestLine.RequestTarget, "/httpbin/"))
	httpHeaders := response.GetDefaultHeaders(0)

	if err != nil {
		handleInternalError(writer, request)
		return
	}

	httpHeaders.Remove("Content-Length")
	httpHeaders.Set("Transfer-Encoding", "chunked")
	httpHeaders.Set("Trailer", "X-Content-SHA256")
	httpHeaders.Set("Trailer", "X-Content-Length")
	httpHeaders.Replace("Content-Type", "text/plain")
	// you can use 32 byte chunks for testing, instead of 1024.
	chunkSize := 1024
	buffer := make([]byte, chunkSize)

	responseBody := []byte{}
	for {
		numberOfBytesRead, err := httpbinResponse.Body.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}

		writer.WriteChunkedBody([]byte(fmt.Sprintf("%x\r\n", numberOfBytesRead))) // hex value, without the first 0x part, outputs 20 instead of 32
		fmt.Printf("Number of bytes read: %d\n", numberOfBytesRead)

		writer.WriteChunkedBody(buffer[:numberOfBytesRead])
		responseBody = append(responseBody, buffer[:numberOfBytesRead]...)

		if numberOfBytesRead < chunkSize {
			writer.WriteChunkedBodyDone()
			break
		}
	}

	sha256Hash := sha256.Sum256(responseBody)
	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", toString(sha256Hash[:]))
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(responseBody)))
	writer.WriteTrailers(*trailers)
}

func toString(bytes []byte) string {
	output := ""
	for _, b := range bytes {
		output += fmt.Sprintf("%02x", b)
	}
	return output
}

func writeStatusHeadersAndBody(writer *response.Writer, status response.StatusCode, headers *headers.Headers, body []byte) {
	writer.WriteStatusLine(status)
	writer.WriteHeaders(*headers)
	writer.WriteBody(body)
}

func handleVideoRequest(writer *response.Writer, request *request.HttpRequest) {
	headers := response.GetDefaultHeaders(0)
	headers.Replace("Content-Type", "video/mp4")

	file, err := os.ReadFile("assets/test.mp4")
	if err != nil {
		log.Fatal(err)
	}

	status := response.StatusOK
	body := file
	headers.Replace("Content-Length", fmt.Sprintf("%d", len(file)))

	writeStatusHeadersAndBody(writer, status, headers, body)
}
