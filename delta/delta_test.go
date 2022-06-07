package delta

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/cool-develope/rdiff-go/chunker"
	"github.com/cool-develope/rdiff-go/signature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var golden = []struct {
	in         string
	out        string
	windowSize uint32
}{
	// {"ab", "abc", 2},
	// {"abcabc", "abceabc", 3},
	// {"abcdef", "abde", 1},
	// {"abcabc", "acabc", 2},
	// {"abcabcabcabcabcabc", "abcabcxabcabc", 3},
	// {"abcabcabcabcabcabc", "abcabcxabcdefabc", 2},
	// {strings.Repeat("a", 1e5), strings.Repeat("a", 1e5) + "x", 64},
	// {strings.Repeat("a", 1e5), "x" + strings.Repeat("a", 1e5), 64},
	// {strings.Repeat("a", 2e3), strings.Repeat("a", 1e3) + "x" + strings.Repeat("a", 1e3), 64},
}

func patch(origin string, deltas []Delta, windowSize uint32) string {
	out := ""
	for _, delta := range deltas {
		st := (delta.blockIndex - 1) * windowSize
		en := delta.blockIndex * windowSize
		if en > uint32(len(origin)) {
			en = uint32(len(origin))
		}

		switch delta.deltaType {
		case BlockKeeped:
			out += origin[st:en]
		case BlockUpdated:
			out += string(delta.updatedBytes)
		case BlockAdded:
			out += string(delta.updatedBytes)
		}
	}

	return out
}

func TestGolden(t *testing.T) {
	for _, g := range golden {
		cin := chunker.NewChunker(strings.NewReader(g.in), g.windowSize)
		sig, err := signature.GetSignature(cin, g.windowSize)
		require.NoError(t, err)

		cout := chunker.NewChunker(strings.NewReader(g.out), g.windowSize)
		deltas, err := GetDelta(sig, cout)
		require.NoError(t, err)

		out := patch(g.in, deltas, g.windowSize)
		assert.Equal(t, out, g.out)
	}
}

func TestRandom(t *testing.T) {
	for i := 0; i < 1000; i++ {
		var totalBytes = rand.Uint32()>>20 + 1000
		var windowSize = uint32(math.Sqrt(float64(totalBytes) / 2))

		var srcBuf, targetBuf bytes.Buffer
		_, err := io.CopyN(&srcBuf, rand.New(rand.NewSource(time.Now().UnixNano())), int64(totalBytes))
		require.NoError(t, err)
		src := srcBuf.Bytes()
		targetBuf.Write(src)

		cin := chunker.NewChunker(&srcBuf, windowSize)
		sig, err := signature.GetSignature(cin, windowSize)
		require.NoError(t, err)

		// create 10% of difference by appending new random data
		newBytes := totalBytes / 10
		targetBuf.Truncate(int(totalBytes - newBytes))
		_, err = io.CopyN(&targetBuf, rand.New(rand.NewSource(time.Now().UnixNano())), int64(newBytes))
		require.NoError(t, err)
		target := targetBuf.Bytes()
		cout := chunker.NewChunker(&targetBuf, windowSize)
		deltas, err := GetDelta(sig, cout)
		require.NoError(t, err)

		out := patch(string(src), deltas, windowSize)
		assert.Equal(t, out, string(target))

		// random shuffle
		targetBuf.Reset()
		copy(target, src)
		shuffleCount := totalBytes / 50
		for i := 0; i < int(shuffleCount); i++ {
			j := rand.Intn(int(totalBytes))
			target[0], target[j] = target[j], target[0]
		}
		targetBuf.Write(target)

		cout = chunker.NewChunker(&targetBuf, windowSize)
		deltas, err = GetDelta(sig, cout)
		require.NoError(t, err)

		out = patch(string(src), deltas, windowSize)
		assert.Equal(t, out, string(target))
	}
}
