package chunker

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var golden = []struct {
	out        [][]byte
	in         string
	windowSize uint32
}{
	{[][]byte{{97}}, "a", 1},
	{[][]byte{{97, 98}}, "ab", 2},
	{[][]byte{{97, 98, 99}}, "abc", 3},
	{[][]byte{{97, 98, 99, 100}}, "abcd", 4},
	{[][]byte{{97, 98}, {99, 100}, {101, 102}}, "abcdef", 2},
	{[][]byte{{97, 98, 99}, {97, 98, 99}}, "abcabc", 3},
	{[][]byte{{97, 98}, {99, 97}, {98, 99}, {97, 98}, {99, 97}, {98, 99}}, "abcabcabcabc", 2},
	{[][]byte{{97, 98, 99}, {100, 101, 102}, {97, 98, 99}}, "abcdefabc", 3},
}

func TestGolden(t *testing.T) {
	for _, g := range golden {
		r := strings.NewReader(g.in + "x")
		c := NewChunker(r, g.windowSize)
		for _, buf := range g.out {
			data, err := c.Next()
			require.NoError(t, err)
			assert.Equal(t, data, buf)
		}
		char, err := c.NextChar()
		require.NoError(t, err)
		assert.Equal(t, char, uint8(0x78))

		_, err = c.Next()
		require.EqualError(t, err, ErrEOF.Error())
	}
}
