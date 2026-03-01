package request

import (
	"fmt"
	"io"
	"testing"

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
	// TODO: this is one way of handling the small (i.e. 3) buffer size in RequestFromReader, see what to do with this.
	if len(data) == 0 {
		return 0, fmt.Errorf("The length of sent data is 0!")
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

	// TODO: add more edge cases tests that you can think of
	// Good POST RequestLine, TODO: probably add body
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

	// Invalid method (out of order) RequestLine
	testChunkReader = &chunkReader{
		data:                 "POST HTTP/1.1 /pizza\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 100000,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.Error(t, err)

	// Invalid version in RequestLine
	testChunkReader = &chunkReader{
		data:                 "POST /pizza HTTP/3\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 42000,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.Error(t, err)

	// Method not capital letters error
	testChunkReader = &chunkReader{
		data:                 "post /pizza HTTP/3\r\nHost: localhost:42069\r\nUser-Agent: curl/8.5\r\nAccept: */*\r\n\r\n",
		numberOfBytesPerRead: 7000,
	}
	reader, err = RequestFromReader(testChunkReader)
	require.Error(t, err)
}
