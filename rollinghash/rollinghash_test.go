package rollinghash

import (
	"hash"
	"hash/adler32"
	"strings"
	"testing"
)

const defaultWindowSize = 64

// Stolen from hash/adler32
var golden = []struct {
	out uint32
	in  string
}{
	//{0x00000001, ""}, // panics
	{0x00620062, "a"},
	{0x012600c4, "ab"},
	{0x024d0127, "abc"},
	{0x03d8018b, "abcd"},
	{0x05c801f0, "abcde"},
	{0x081e0256, "abcdef"},
	{0x0adb02bd, "abcdefg"},
	{0x0e000325, "abcdefgh"},
	{0x118e038e, "abcdefghi"},
	{0x158603f8, "abcdefghij"},
	{0x211297c8, strings.Repeat("\xff", 5548) + "8"},
	{0xbaa198c8, strings.Repeat("\xff", 5549) + "9"},
	{0x553499be, strings.Repeat("\xff", 5550) + "0"},
	{0xf0c19abe, strings.Repeat("\xff", 5551) + "1"},
	{0x8d5c9bbe, strings.Repeat("\xff", 5552) + "2"},
	{0x2af69cbe, strings.Repeat("\xff", 5553) + "3"},
	{0xc9809dbe, strings.Repeat("\xff", 5554) + "4"},
	{0x69189ebe, strings.Repeat("\xff", 5555) + "5"},
	{0x86af0001, strings.Repeat("\x00", 1e5)},
	{0x79660b4d, strings.Repeat("a", 1e5)},
	{0x110588ee, strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 1e4)},
}

// Sum32ByWriteAndRoll computes the sum by prepending the input slice with
// a '\0', writing the first bytes of this slice into the sum, then
// sliding on the last byte and returning the result of Sum32
func Sum32ByWriteAndRoll(b []byte) uint32 {
	q := []byte("\x00")
	q = append(q, b...)
	roll := New(defaultWindowSize)
	roll.Write(q[:len(q)-1]) //nolint
	roll.Roll(q[len(q)-1])
	return roll.Sum32()
}

func TestGolden(t *testing.T) {
	for _, g := range golden {
		in := g.in

		// We test the classic implementation
		p := []byte(g.in)
		classic := hash.Hash32(adler32.New())
		classic.Write(p) //nolint
		if got := classic.Sum32(); got != g.out {
			t.Errorf("classic implementation: for %q, expected 0x%x, got 0x%x", in, g.out, got)
			continue
		}

		if got := Sum32ByWriteAndRoll(p); got != g.out {
			t.Errorf("rolling implementation: for %q, expected 0x%x, got 0x%x", in, g.out, got)
			continue
		}
	}
}
