package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParsing(t *testing.T) {
	// Test: Invalid spacing header
	headers := NewHeaders()
	data := []byte("       Host :       localhost:42069      \r\n Accept: video/*\r\n\r\n")
	bytesParsed, done, err := headers.Parse(data)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrorWhitespaceBetweenColonAndKey)
	assert.Equal(t, 0, bytesParsed)
	assert.False(t, done)

	// Test: Valid spacing in multiple headers
	headers = NewHeaders()
	data = []byte("       Host:       localhost:42069      \r\n Accept: video/*\r\n\r\n")
	bytesParsed, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "localhost:42069", headers.Get("host"))
	assert.Equal(t, "video/*", headers.Get("Accept"))
	assert.Equal(t, "video/*", headers.Get("accept"))
	assert.Equal(t, 62, bytesParsed) // \r and  \n are both considered 1 byte
	assert.True(t, done)

	// Test: Case insensitive keys in the header
	headers = NewHeaders()
	data = []byte("       host:       localhost:42069      \r\n accept: video/*\r\n\r\n")
	bytesParsed, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "localhost:42069", headers.Get("host"))
	assert.Equal(t, "video/*", headers.Get("Accept"))
	assert.Equal(t, "video/*", headers.Get("accept"))
	assert.Equal(t, 62, bytesParsed)
	assert.True(t, done)

	// Test: Header contains invalid characters
	headers = NewHeaders()
	data = []byte("       hostž:       localhost:42069      \r\n accept: video/*\r\n\r\n")
	bytesParsed, done, err = headers.Parse(data)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrorHeaderContainsInvalidCharacters)
	assert.Equal(t, 0, bytesParsed)
	assert.False(t, done)

	// Test: Header contains multiple values for the same key
	headers = NewHeaders()
	data = []byte("       host:       localhost:42069      \r\n accept: video/*\r\n accept: video/*\r\n\r\n")
	bytesParsed, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 80, bytesParsed)
	assert.Equal(t, "localhost:42069", headers.Get("HOST"))
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "video/*, video/*", headers.Get("ACCEPT"))
	assert.Equal(t, "video/*, video/*", headers.Get("Accept"))
	assert.True(t, done)
}
