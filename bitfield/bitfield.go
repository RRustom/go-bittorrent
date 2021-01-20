package bitfield

type Bitfield []byte

func (bf Bitfield) HasPiece(index int) bool {
  byteIndex := index / 8
  offset := index % 8
  if byteIndex < 0 || byteIndex >= len(bf) {
    return false
  }
  return bf[byteIndex] >> (7-offset)&1 != 0
}

// set a bit in the bitfield
func (bf Bitfield) SetPiece(index int) {
  byteIndex := index / 8
  offset := index % 8

  if byteIndex < 0 || byteIndex >= len(bf) {
    return  // silently discard
  }

  bf[byteIndex] |= 1 << (7 - offset)
}
