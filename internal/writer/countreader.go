package writer

import "io"

// CountingReader wraps an io.Reader and tracks the number of bytes read.
type CountingReader struct {
	reader    io.Reader
	BytesRead int64
}

// NewCountingReader creates a new CountingReader wrapping the given reader.
func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{reader: r}
}

// Read implements io.Reader.
func (cr *CountingReader) Read(p []byte) (int, error) {
	n, err := cr.reader.Read(p)
	cr.BytesRead += int64(n)
	return n, err
}
