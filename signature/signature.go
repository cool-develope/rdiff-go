package signature

import (
	"crypto/md5"

	"github.com/cool-develope/rdiff-go/chunker"
	"github.com/cool-develope/rdiff-go/rollinghash"
)

// Signature is a struct to represent the checkSum
type Signature struct {
	WindowSize uint32
	StrongSigs [][]byte         // slice of MD5 hash values
	Weak2Sigs  map[uint32][]int // the map to reference the block indexes for the specific weak hash
}

// GetSignature creates the signature from the stream data
func GetSignature(c *chunker.Chunker, windowSize uint32) (*Signature, error) {
	var sig Signature
	sig.WindowSize = windowSize
	sig.Weak2Sigs = make(map[uint32][]int)

	rHash := rollinghash.New(windowSize)
	sHash := md5.New()

	for {
		buf, err := c.Next()
		if err != nil {
			if err == chunker.ErrEOF {
				break
			}
			return nil, err
		}

		rHash.Reset()
		_, err = rHash.Write(buf)
		if err != nil {
			return nil, err
		}
		rSum := rHash.Sum32()

		sHash.Reset()
		_, err = sHash.Write(buf)
		if err != nil {
			return nil, err
		}
		sSum := sHash.Sum(nil)

		if indexes, found := sig.Weak2Sigs[rSum]; found {
			sig.Weak2Sigs[rSum] = append(indexes, len(sig.StrongSigs))
		} else {
			sig.Weak2Sigs[rSum] = []int{len(sig.StrongSigs)}
		}

		sig.StrongSigs = append(sig.StrongSigs, sSum)
	}

	return &sig, nil
}
