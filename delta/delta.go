package delta

import (
	"bytes"
	"crypto/md5"

	"github.com/balena-os/circbuf"
	"github.com/cool-develope/rdiff-go/chunker"
	"github.com/cool-develope/rdiff-go/rollinghash"
	"github.com/cool-develope/rdiff-go/signature"
)

// Type is a type definition for enum
type Type uint8

const (
	// BlockUpdated means the current block is updated in new version
	BlockUpdated Type = iota
	// BlockRemoved means the current block is removed
	BlockRemoved
	// BlockAdded means new block is added after the current block
	BlockAdded
	// BlockKeeped means this block is keeped
	BlockKeeped
)

// Delta is a struct to represent the difference
type Delta struct {
	deltaType    Type
	blockIndex   uint32
	updatedBytes []byte
}

type updatedElement struct {
	blockIndex int
	char       byte
	keeped     bool
}

// GetDelta returns the diffing
func GetDelta(sig *signature.Signature, c *chunker.Chunker) ([]Delta, error) {
	rHash := rollinghash.New(sig.WindowSize)
	sHash := md5.New()
	pos := -1
	prevChar := byte(0)
	block, _ := circbuf.NewBuffer(int64(sig.WindowSize))
	matched := true
	updates := make([]updatedElement, 0)

	for {
		if matched {
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

			block.Reset()
			_, err = block.Write(buf)
			if err != nil {
				return nil, err
			}
		} else {
			char, err := c.NextChar()
			if err != nil {
				if err == chunker.ErrEOF {
					for _, char := range block.Bytes() {
						updates = append(updates, updatedElement{
							blockIndex: pos,
							char:       char,
						})
					}
					break
				}
				return nil, err
			}

			prevChar, err = block.Get(0)
			if err != nil {
				return nil, err
			}

			updates = append(updates, updatedElement{
				blockIndex: pos,
				char:       prevChar,
			})

			err = block.WriteByte(char)
			if err != nil {
				return nil, err
			}
			rHash.Roll(char)
		}

		matched = false
		rSum := rHash.Sum32()
		if indexes, found := sig.Weak2Sigs[rSum]; found {
			sHash.Reset()
			_, err := sHash.Write(block.Bytes())
			if err != nil {
				return nil, err
			}
			sSum := sHash.Sum(nil)

			for _, index := range indexes {
				if index > pos && bytes.Equal(sig.StrongSigs[index], sSum) {
					matched = true
					pos = index
					updates = append(updates, updatedElement{
						blockIndex: index,
						keeped:     true,
					})
					break
				}
			}
		}
	}

	// refactor the result
	results := make([]Delta, 0)
	prevIndex := -1
	temp := make([]byte, 0)

	for _, ele := range updates {
		if ele.keeped {
			results = append(results, compose(temp, prevIndex, ele.blockIndex, sig.WindowSize)...)
			results = append(results, Delta{
				deltaType:  BlockKeeped,
				blockIndex: uint32(ele.blockIndex + 1),
			})
			temp = temp[:0]
		} else {
			temp = append(temp, ele.char)
		}
		prevIndex = ele.blockIndex
	}

	results = append(results, compose(temp, prevIndex, len(sig.StrongSigs), sig.WindowSize)...)
	return results, nil
}

func compose(temp []byte, prevIndex, curIndex int, windowSize uint32) []Delta {
	deltas := make([]Delta, 0)
	pos := 0

	for i := prevIndex + 1; i < curIndex; i++ {
		if pos >= len(temp) {
			deltas = append(deltas, Delta{
				deltaType:  BlockRemoved,
				blockIndex: uint32(i + 1),
			})
		} else {
			end := pos + int(windowSize)
			if end > len(temp) {
				end = len(temp)
			}
			block := make([]byte, end-pos)
			copy(block, temp[pos:end])
			deltas = append(deltas, Delta{
				deltaType:    BlockUpdated,
				blockIndex:   uint32(i + 1),
				updatedBytes: block,
			})
			pos = end
		}
	}

	for ; pos < len(temp); pos += int(windowSize) {
		end := pos + int(windowSize)
		if end > len(temp) {
			end = len(temp)
		}
		block := make([]byte, end-pos)
		copy(block, temp[pos:end])
		deltas = append(deltas, Delta{
			deltaType:    BlockAdded,
			blockIndex:   uint32(curIndex),
			updatedBytes: block,
		})
	}

	return deltas
}
