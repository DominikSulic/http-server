package response

import (
	"fmt"
	"io"

	"http-server/internal/headers"
)

// These are fake go enums.
type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// maps the given status code to the correct reason phrase, if its supported. Any other code should just leave the reason phrase blank.
func WriteStatusLine(writer io.Writer, statusCode StatusCode) error {
	statusLine := "HTTP/1.1 "

	switch statusCode {
	case StatusOK:
		statusLine += "200 OK\r\n"
		break

	case StatusBadRequest:
		statusLine += "400 Bad Request\r\n"
		break

	case StatusInternalServerError:
		statusLine += "500 Internal Server Error\r\n"
		break

	default:
		statusLine += string(statusCode)
	}

	_, err := writer.Write([]byte(statusLine))

	return err
}

// Sets the following headers that you always wanna include in responses:
// content-length, connection (close because we're not doing keep-alive connections yet)
func GetDefaultHeaders(contentLength int) *headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", fmt.Sprintf("%d", contentLength))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}

func WriteHeaders(writer io.Writer, headers *headers.Headers) error {
	byteArray := []byte{}

	headers.ForEach(func(key, value string) {
		byteArray = fmt.Appendf(byteArray, "%s: %s\r\n", key, value)
	})

	byteArray = fmt.Append(byteArray, "\r\n")
	_, err := writer.Write(byteArray)

	return err
}
