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
	host, ok := headers.Get("Host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)
	host, ok = headers.Get("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)
	accept, ok := headers.Get("Accept")
	assert.True(t, ok)
	assert.Equal(t, "video/*", accept)
	accept, ok = headers.Get("accept")
	assert.True(t, ok)
	assert.Equal(t, "video/*", accept)
	assert.Equal(t, 62, bytesParsed) // \r and  \n are both considered 1 byte
	assert.True(t, done)

	// Test: Case insensitive keys in the header
	headers = NewHeaders()
	data = []byte("       host:       localhost:42069      \r\n accept: video/*\r\n\r\n")
	bytesParsed, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	host, ok = headers.Get("Host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)
	host, ok = headers.Get("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)
	accept, ok = headers.Get("Accept")
	assert.True(t, ok)
	assert.Equal(t, "video/*", accept)
	accept, ok = headers.Get("accept")
	assert.True(t, ok)
	assert.Equal(t, "video/*", accept)
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
	host, ok = headers.Get("HOST")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)
	host, ok = headers.Get("Host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)
	accept, ok = headers.Get("ACCEPT")
	assert.True(t, ok)
	assert.Equal(t, "video/*, video/*", accept)
	accept, ok = headers.Get("Accept")
	assert.True(t, ok)
	assert.Equal(t, "video/*, video/*", accept)
	assert.True(t, done)
}
