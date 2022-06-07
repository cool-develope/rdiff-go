# rdiff-go
`rdiff` is a utility for efficiency file diffing algorithm. When comparing original and an updated version of an input, it should return a description (`delta`) which can be used to upgrade an original version of the file into the new file. The description provides information of the chunks which:

- Can be reused from the original file
- Have been added or modified and thus would need to be synchronized

## Assumption
To compare original and updated input, we are using rolling hash algorithm based on `adler32` as a weak hash and `MD5` as a strong hash. 

**Signature Format**

```go
type Signature struct {
    windowSize uint32
    strongSigs [][]byte // slice of MD5 hash values
    weak2Sigs map[uint32][]int // the map to reference the block indexes for the specific weak hash
}
```

**Delta Format**

```go
type DeltaType uint8

const (
    Block_Updated DeltaType = iota
    Block_Removed
    Block_Added
    Block_Keeped
)

type Delta struct {
    deltaType DeltaType
    blockIndex uint32
    updatedBytes []byte
}

```

- `Block_Updated, 5, [123, 23, 43]` : 5th block of the original input is replaced by `[123, 23, 43]`
- `Block_Removed, 6, nil` : 6th block of the original input is removed
- `Block_Added, 7, [123, 23, 43]` : `[123, 23, 43]` is added after the 7th block of the original input

## Further More
- use different checksum algorithms such as `rabinkarp64`, `buzhash64` instead of `adler32`
- `Delta` is designed like one way (increasing cursor)
- it is not matching the tail of input when it is not completely divided by `windowSize`
- `Delta` can be big enough, need to use the file stream to store the result