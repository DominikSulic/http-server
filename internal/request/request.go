package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type parserState string

const (
	StateInitialized parserState = "init"
	StateDone        parserState = "done"
	StateError       parserState = "error"
)

var (
	ErrorMalformedRequestLine    = fmt.Errorf("Malformed request-line!")
	ErrorMethodNotCapitalLetters = fmt.Errorf("Request method is not written using only capital letters!")
	ErrorHttpRequestInErrorState = fmt.Errorf("Request is in error state!")
	Separator                    = []byte("\r\n")
)

// if its a request, not a response, the start line is referred to as the "request-line" and has a specific format.
// An example of the requestline: GET /pizza HTTP/1.1
type RequestLine struct {
	HttpVersion   string // HTTP-name "/" DIGIT "." DIGIT
	RequestTarget string // %s "HTTP"
	Method        string // method SP request-target SP HTTP-version
}

type HttpRequest struct {
	RequestLine RequestLine
	state       parserState
}

func (requestLine *RequestLine) ValidHTTPVersion() bool {
	return requestLine.HttpVersion == "HTTP/1.1"
}

func newHttpRequest() *HttpRequest {
	return &HttpRequest{
		state: StateInitialized,
	}
}

// HttpRequest respresents a full parsed HTTP request, 	dont do ReadAll, do a for loop and read bit by bit, loading in everything at once is not that good.
func RequestFromReader(reader io.Reader) (*HttpRequest, error) {
	httpRequest := newHttpRequest()

	// NOTE: buffer could get overrun, anything exceeding 1k - the header or the body for instance
	buffer := make([]byte, 4096)
	bufferLength := 0

	for !httpRequest.done() {

		numberOfBytesRead, err := reader.Read(buffer[bufferLength:])
		// TODO: see what to do with these
		if err != nil {
			return nil, errors.Join(fmt.Errorf("Unable to read from the reader: "), err)
		}

		bufferLength += numberOfBytesRead

		numberOfBytesParsed, err := httpRequest.parse(buffer[:bufferLength])
		if err != nil {
			return nil, err
		}

		copy(buffer, buffer[numberOfBytesParsed:bufferLength]) // the 1st parameter is the destination
		bufferLength -= numberOfBytesParsed
	}

	return httpRequest, nil
}

func (httpRequest *HttpRequest) parse(data []byte) (int, error) {
	read := 0

	// for loop cause you could get a piece of really huge data containing several state transitions - so you wanna be able to for loop
	// this outer thing is labeling - its one way of returning from a deeply nested item... labeling things usually is eh ?
outer:
	for {
		switch httpRequest.state {
		case StateError:
			return 0, ErrorHttpRequestInErrorState

		case StateInitialized:
			requestLine, numberOfBytesProcessed, err := parseRequestLine(data[read:])
			if err != nil {
				httpRequest.state = StateError
				return 0, err
			}

			if numberOfBytesProcessed == 0 {
				break outer
			}

			httpRequest.RequestLine = *requestLine
			read += numberOfBytesProcessed

			httpRequest.state = StateDone // TODO: figure out why this makes sense

		case StateDone:
			break outer
		}
	}

	return read, nil
}

func (httpRequest *HttpRequest) done() bool {
	return httpRequest.state == StateDone || httpRequest.state == StateError // technically done if an error occurred
}

// int is the number of bytes processed, string is a byte array effectively so we can change the string to []byte
func parseRequestLine(data []byte) (*RequestLine, int, error) {
	index := bytes.Index(data, Separator)

	if index == -1 {
		return nil, 0, nil
	}

	requestLineData := data[:index]
	read := index + len(Separator) // Do not include the separator!!

	requestLineParts := bytes.Split(requestLineData, []byte(" ")) // RFC 9112 says its a single space between the parts

	if len(requestLineParts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpVersionParts := bytes.Split(requestLineParts[2], []byte("/"))

	if len(httpVersionParts) != 2 || string(httpVersionParts[0]) != "HTTP" || string(httpVersionParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	for _, charNumber := range requestLineParts[0] {
		if charNumber > 90 || charNumber < 65 {
			return nil, index, ErrorMethodNotCapitalLetters
		}
	}

	return &RequestLine{Method: string(requestLineParts[0]), RequestTarget: string(requestLineParts[1]), HttpVersion: string(httpVersionParts[1])}, read, nil
}
