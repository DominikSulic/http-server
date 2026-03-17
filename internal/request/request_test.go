package request

import (
	"io"
	"testing"

	"http-server/internal/headers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data                 string
	numberOfBytesPerRead int
	position             int
}

// Read reads up to len(p) or numberOfBytesPerRead bytes from the string, useful for simulating reading a variable number of bytes per chunk
func (chunkReader *chunkReader) Read(data []byte) (n int, err error) {
	// this is one way of handling the small (i.e. 3) buffer size in RequestFromReader
	if len(data) == 0 {
		return 0, nil
	}

	if chunkReader.position >= len(chunkReader.data) {
		return 0, io.EOF
	}

	endIndex := chunkReader.position + chunkReader.numberOfBytesPerRead
	if endIndex > len(chunkReader.data) {
		endIndex = len(chunkReader.data)
	}

	n = copy(data, chunkReader.data[chunkReader.position:endIndex])
	chunkReader.position += n
	if n > chunkReader.numberOfBytesPerRead {
		n = chunkReader.numberOfBytesPerRead
		chunkReader.position -= n - chunkReader.numberOfBytesPerRead
	}

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: good GET request line
	testChunkReader := &chunkReader{
		data:                 "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 3,
	}
	reader, err := RequestFromReader(testChunkReader)
	require.NoError(t, err)
	require.NotNil(t, reader)
	assert.Equal(t, "GET", reader.RequestLine.Method)
	assert.Equal(t, "/", reader.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", reader.RequestLine.HttpVersion)

	// Test: good GET request line with path
	testChunkReader = &chunkReader{
		data:                 "GET /pizza HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 1,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.NoError(t, err)
	require.NotNil(t, reader)
	assert.Equal(t, "GET", reader.RequestLine.Method)
	assert.Equal(t, "/pizza", reader.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", reader.RequestLine.HttpVersion)

	// Test: invalid number of parts in request line
	testChunkReader = &chunkReader{
		data:                 "/pizza HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 200000000000000,
	}
	_, err = RequestFromReader(testChunkReader)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrorMalformedRequestLine)

	// Good POST RequestLine
	testChunkReader = &chunkReader{
		data:                 "POST /pizza HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 80,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.NoError(t, err)
	require.NotNil(t, reader)
	assert.Equal(t, "POST", reader.RequestLine.Method)
	assert.Equal(t, "/pizza", reader.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", reader.RequestLine.HttpVersion)

	// Invalid number of parts in request line
	testChunkReader = &chunkReader{
		data:                 "POST /pizza HTTP/1.1 /randomThingToFailTheTest\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 100,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrorMalformedRequestLine)

	// Invalid method (out of order) RequestLine
	testChunkReader = &chunkReader{
		data:                 "POST HTTP/1.1 /pizza\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 100000,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrorMalformedRequestLine)

	// Invalid version in RequestLine
	testChunkReader = &chunkReader{
		data:                 "POST /pizza HTTP/3\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 42000,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrorMalformedRequestLine)

	// Method not capital letters error
	testChunkReader = &chunkReader{
		data:                 "post /pizza HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 7000,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrorMethodNotCapitalLetters)
}

func TestRequestHeaders(t *testing.T) {
	// Test : standard headers
	testChunkReader := &chunkReader{
		data:                 "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 10000,
	}
	reader, err := RequestFromReader(testChunkReader)
	require.NoError(t, err)
	require.NotNil(t, reader)
	host, ok := reader.Headers.Get("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)
	userAgent, ok := reader.Headers.Get("user-agent")
	assert.True(t, ok)
	assert.Equal(t, "curl/8.5", userAgent)
	accept, ok := reader.Headers.Get("accept")
	assert.True(t, ok)
	assert.Equal(t, "*/*", accept)

	// Test : malformed header
	testChunkReader = &chunkReader{
		data:                 "GET / HTTP/1.1\r\nHost : localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 42,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.Error(t, err)
	require.ErrorIs(t, err, headers.ErrorWhitespaceBetweenColonAndKey)

	// Test : duplicate headers
	testChunkReader = &chunkReader{
		data:                 "GET / HTTP/1.1\r\n       host:       localhost:42069      \r\n accept: video/*\r\n accept: video/*\r\n\r\n",
		numberOfBytesPerRead: 7000,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.NoError(t, err)
	require.NotNil(t, reader)
	host, ok = reader.Headers.Get("host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)
	accept, ok = reader.Headers.Get("accept")
	assert.True(t, ok)
	assert.Equal(t, "video/*, video/*", accept)

	// Test : empty headers
	testChunkReader = &chunkReader{
		data:                 "GET / HTTP/1.1\r\n\r\n",
		numberOfBytesPerRead: 13,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.NoError(t, err)
	require.NotNil(t, reader)
}

func TestRequestBody(t *testing.T) {
	// Test: standard body
	testChunkReader := &chunkReader{
		data:                 "POST /requestPizza HTTP/1.1\r\nHost: localhost:42069\r\nContent-Length: 13\r\n\r\nhello world!\n",
		numberOfBytesPerRead: 10000000000,
	}
	reader, err := RequestFromReader(testChunkReader)
	require.NoError(t, err)
	require.NotNil(t, reader)
	assert.Equal(t, "hello world!\n", string(reader.Body))

	// Test: Body shorter than sent Content-Length
	testChunkReader = &chunkReader{
		data:                 "POST /requestPizza HTTP/1.1\r\nHost: localhost:42069\r\nContent-Length: 15\r\n\r\nhello world!\n",
		numberOfBytesPerRead: 10000000000,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrorHttpBodyNotEqualToContentLength)

	// Test: no content-length but body exists -> shouldnt error
	testChunkReader = &chunkReader{
		data:                 "POST /requestPizza HTTP/1.1\r\nHost: localhost:42069\r\n\r\nHELLO WORLD!\n",
		numberOfBytesPerRead: 10000000000,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.NoError(t, err)
	require.NotNil(t, reader)
	assert.Equal(t, "", string(reader.Body))

	// Test: empty body, 0 content-length
	testChunkReader = &chunkReader{
		data:                 "POST /requestPizza HTTP/1.1\r\nHost: localhost:42069\r\nContent-Length: 0\r\n\r\n\n",
		numberOfBytesPerRead: 10000000000,
	}
	require.NoError(t, err)
	require.NotNil(t, reader)
	assert.Equal(t, "", string(reader.Body))
}
