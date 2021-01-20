package handshake

import (
  "io"
  "fmt"
)

// a handshake is a special message that a peer uses to identify itself
type Handshake struct {
  Pstr      string
  InfoHash  [20]byte
  PeerID    [20]byte
}

// create a new handshake with the standard pstr
func New(infoHash, peerID [20]byte) *Handshake {
  return &Handshake{
    Pstr:     "BitTorrent protocol",
    InfoHash: infoHash,
    PeerID:   peerID,
  }
}

// serialize the handshake to a buffer
func (h *Handshake) Serialize() []byte {
  buf := make([]byte, len(h.Pstr)+49)
  buf[0] = byte(len(h.Pstr))
  current := 1
  current += copy(buf[current:], h.Pstr)
  current += copy(buf[current:], make([]byte, 8)) // 8 reserved bytes
  current += copy(buf[current:], h.InfoHash[:])
  current += copy(buf[current:], h.PeerID[:])
  return buf
}

// parse a handshake from a stream
func Read(r io.Reader) (*Handshake, error) {
  lengthBuf := make([]byte, 1)
  _, err := io.ReadFull(r, lengthBuf)
  if err != nil {
    return nil, err
  }
  pstrlen := int(lengthBuf[0])

  if pstrlen == 0 {
    err := fmt.Errorf("pstrlen cannot be 0")
    return nil, err
  }

  handshakeBuf := make([]byte, 48+pstrlen)
  _, err = io.ReadFull(r, handshakeBuf)
  if err != nil {
    return nil, err
  }

  var infoHash, peerID [20]byte

  copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
  copy(peerID[:], handshakeBuf[pstrlen+8+20:])

  h := Handshake{
    Pstr: string(handshakeBuf[0:pstrlen]),
    InfoHash: infoHash,
    PeerID: peerID,
  }

  return &h, nil
}
