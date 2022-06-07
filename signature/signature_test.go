package signature

import (
	"strings"
	"testing"

	"github.com/cool-develope/rdiff-go/chunker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint
var golden = []struct {
	out        map[uint32][]int
	in         string
	windowSize uint32
}{
	{map[uint32][]int{0x620062: []int{0}}, "a", 1},
	{map[uint32][]int{0x12600c4: []int{0}}, "ab", 2},
	{map[uint32][]int{0x24d0127: []int{0}}, "abc", 3},
	{map[uint32][]int{0x3d8018b: []int{0}}, "abcd", 4},
	{map[uint32][]int{0x660066: []int{2}, 0x12600c4: []int{0}, 0x12c00c8: []int{1}}, "abcde", 2},
	{map[uint32][]int{0x24d0127: []int{0, 1}}, "abcabc", 3},
	{map[uint32][]int{0x2650133: []int{1}, 0x3d8018b: []int{0}}, "abcdefg", 4},
	{map[uint32][]int{0x26b0136: []int{1}, 0x5c801f0: []int{0}}, "abcdefgh", 5},
	{map[uint32][]int{0x12600c4: []int{0, 3}, 0x12900c5: []int{1, 4}, 0x12900c6: []int{2, 5}}, "abcabcabcabc", 2},
	{map[uint32][]int{0x24d0127: []int{0, 2}, 0x25f0130: []int{1}}, "abcdefabc", 3},
}

func TestGolden(t *testing.T) {
	for _, g := range golden {
		b := strings.NewReader(g.in)
		c := chunker.NewChunker(b, g.windowSize)
		sig, err := GetSignature(c, g.windowSize)
		require.NoError(t, err)

		assert.Equal(t, g.out, sig.Weak2Sigs)
	}
}
