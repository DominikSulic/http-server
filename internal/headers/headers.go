package headers

import (
	"bytes"
	"fmt"
	"strings"
)

// the package is going to be used for both parsing requests and sending responses.

type Headers struct {
	headers map[string]string
}

var CRLF []byte = []byte("\r\n")

var (
	ErrorWhitespaceBetweenColonAndKey    = fmt.Errorf("There is a whitespace character between the colon and the key in headers!")
	ErrorMalformedHeader                 = fmt.Errorf("Malformed header!")
	ErrorHeaderContainsInvalidCharacters = fmt.Errorf("There is an invalid character in the header!")
)

func NewHeaders() *Headers {
	return &Headers{headers: make(map[string]string)}
}

func (headers *Headers) Get(name string) (string, bool) {
	string, ok := headers.headers[strings.ToLower(name)]
	return string, ok
}

func (headers *Headers) Replace(name string, value string) {
	headerName := strings.ToLower(name)

	headers.headers[headerName] = value
}

func (headers *Headers) Set(name string, value string) {
	headerName := strings.ToLower(name)

	if headerValue, found := headers.headers[name]; found {
		headers.headers[headerName] = fmt.Sprintf("%s, %s", headerValue, value)
	} else {
		headers.headers[headerName] = value
	}
}

// done this way not to expose the internals of the module to the outside
func (headers *Headers) ForEach(cb func(key, value string)) {
	for key, value := range headers.headers {
		cb(key, value)
	}
}

/*
 * A header name (the key only, not the value itself) can contain only:
 * uppercase and lowercase letters A-Z
 * digits 0-9
 * ! # $ % & ' * + - . ^ _ ` | ~
 * and is at least of length 1
 */
func checkHeaderKeyForInvalidCharacters(key []byte) bool {
	validCharacters := []string{"!", "#", "$", "%", "&", "'", "*", "+", "-", ".", "^", "_", "`", "|", "~"}

	if len(key) < 1 {
		return false
	}

	containsInvalidCharacters := false

	// it always returns the index and then the element itself, _ is saying you dont care about the index
	for _, char := range key {

		// digits, uppercase a-z, lowercase a-z
		if (char >= 48 && char <= 57) || (char >= 65 && char <= 90) || (char >= 97 && char <= 122) {
			continue
		}

		validSpecialCharacter := false

		for _, validChar := range validCharacters {
			if string(char) == string(validChar) {
				validSpecialCharacter = true
			}
		}

		if !validSpecialCharacter {
			containsInvalidCharacters = true
			break
		}
	}

	return containsInvalidCharacters
}

func parseHeader(header []byte) (string, string, error) {
	// 2 subslices -> because the value can be localhost:42069 and the second colon would make the bytes.Split method break
	headerParts := bytes.SplitN(header, []byte(":"), 2)

	// the method can return 1 for instance, it 'returns at most n subslices' (2 in this case cause thats the value sent)
	if len(headerParts) != 2 {
		return "", "", ErrorMalformedHeader
	}

	key := bytes.TrimLeft(bytes.ToLower(headerParts[0]), " ") // removes spacing before the key
	value := bytes.TrimSpace(headerParts[1])

	if bytes.HasSuffix(key, []byte(" ")) { // there can be no whitespace between the key and the colon, not sure why not just trim it though
		return "", "", ErrorWhitespaceBetweenColonAndKey
	}

	if checkHeaderKeyForInvalidCharacters(key) {
		return "", "", ErrorHeaderContainsInvalidCharacters
	}

	return string(key), string(value), nil
}

/*
 * Example headers:
 * Host: youtube.com
 * Authorization: Api-Key 69woof
 * Accept: video/*
 * User-Agent: DogFetcher/1.0
 */
func (headers Headers) Parse(data []byte) (bytesParsed int, done bool, err error) {
	bytesRead := 0
	finished := false

	for {

		crLfIndex := bytes.Index(data[bytesRead:], CRLF)

		if crLfIndex == -1 {
			break
		}

		if crLfIndex == 0 { // crLfIndex == 0 means we reached a line that just contains \r\n -> done with parsing the headers.
			bytesRead += len(CRLF)
			finished = true
			break
		}

		key, value, err := parseHeader(data[bytesRead : bytesRead+crLfIndex])
		if err != nil {
			return 0, false, err
		}

		headers.Set(key, value)

		bytesRead += crLfIndex + len(CRLF)
	}

	return bytesRead, finished, nil
}
