package chunker

import (
	"fmt"
	"io"
)

// ErrEOF is an error to represent EOF
var ErrEOF = fmt.Errorf("EOF")

// Chunker is a stream utility to chunk data
type Chunker struct {
	windowSize uint32
	rd         io.Reader
}

// NewChunker returns new chunker object
func NewChunker(rd io.Reader, windowSize uint32) *Chunker {
	return &Chunker{
		windowSize: windowSize,
		rd:         rd,
	}
}

// Next returns the next chunk from the stream
func (c *Chunker) Next() ([]byte, error) {
	buf := make([]byte, 0, c.windowSize)
	n, err := io.ReadFull(c.rd, buf[:cap(buf)])

	if err != nil {
		if err == io.EOF {
			return nil, ErrEOF
		}
		if err != io.ErrUnexpectedEOF {
			return nil, err
		}
	}
	buf = buf[:n]

	return buf, nil
}

// NextChar return the next char
func (c *Chunker) NextChar() (byte, error) {
	buf := make([]byte, 0, 1)
	n, err := io.ReadFull(c.rd, buf[:cap(buf)])

	if err != nil {
		if err == io.EOF {
			return byte(0), ErrEOF
		}

		return byte(0), err
	}
	buf = buf[:n]

	return buf[0], err
}
