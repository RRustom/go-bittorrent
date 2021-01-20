package message

import (
  "encoding/binary"
  "fmt"
  "io"
)

type messageID uint8

const (
  MsgChoke          messageID = 0 // chokes the receiver
  MsgUnchoke        messageID = 1 // unchokes the receiver
  MsgInterested     messageID = 2 // expresses interest in receiving data
  MsgNotInterested  messageID = 3 // expresses disinterest in receiving data
  MsgHave           messageID = 4 // alerts receiver that sender has downloaded a piece
  MsgBitfield       messageID = 5 // encodes which pieces the sender has downloaded
  MsgRequest        messageID = 6 // requests a block of data from the receiver
  MsgPiece          messageID = 7 // delivers a block of data to fulfill a request
  MsgCancel         messageID = 8 // cancels a request
)

type Message struct {
  ID      messageID
  Payload []byte
}

// creates a REQUEST message
func FormatRequest(index, begin, length int) *Message {
  payload := make([]byte, 12)
  binary.BigEndian.PutUint32(payload[0:4], uint32(index))
  binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
  binary.BigEndian.PutUint32(payload[8:12], uint32(length))
  return &Message{ID: MsgRequest, Payload: payload}
}

// creates a HAVE message
func FormatHave(index int) *Message {
  payload := make([]byte, 4)
  binary.BigEndian.PutUint32(payload, uint32(index))
  return &Message{ID: MsgHave, Payload: payload}
}

// parses a PIECE message and copies its payload into a buffer
func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
  if msg.ID != MsgPiece {
    return 0, fmt.Errorf("Expected PIECE (ID %d), got ID %d", MsgPiece, msg.ID)
  }

  if len(msg.Payload) < 8 {
    return 0, fmt.Errorf("Payload too short. %d < 8", len(msg.Payload))
  }

  parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
  if parsedIndex != index {
    return 0, fmt.Errorf("Expected index %d, got %d", index, parsedIndex)
  }

  begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
  if begin >= len(buf) {
    return 0, fmt.Errorf("Begin offset too high. %d >= %d", begin, len(buf))
  }

  data := msg.Payload[8:]
  if begin + len(data) > len(buf) {
    return 0, fmt.Errorf("Data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
  }
  copy(buf[begin:], data)
  return len(data), nil
}

// parses a HAVE message
func ParseHave(msg *Message) (int, error) {
  if msg.ID != MsgHave {
    return 0, fmt.Errorf("Expected HAVE (ID %d), got ID %d", MsgHave, msg.ID)
  }
  if len(msg.Payload) != 4 {
    return 0, fmt.Errorf("Expected payload length 4, got length %d", len(msg.Payload))
  }
  index := int(binary.BigEndian.Uint32(msg.Payload))
  return index, nil
}

// serialize a message into a a buffer of the form
// <length prefix><message ID><payload>
// Intereprets `nil` as keep-alive message
func (m *Message) Serialize() []byte {
  if m == nil {
    return make([]byte, 4)
  }
  length := uint32(len(m.Payload) + 1)
  buf := make([]byte, 4+length)
  binary.BigEndian.PutUint32(buf[0:4], length)
  buf[4] = byte(m.ID)
  copy(buf[5:], m.Payload)
  return buf
}

// parse a message from a stream. Return `nil` on keep-alive message
func Read(r io.Reader) (*Message, error) {
  lengthBuf := make([]byte, 4)
  _, err := io.ReadFull(r, lengthBuf)
  if err != nil {
    return nil, err
  }

  length := binary.BigEndian.Uint32(lengthBuf)
  // keep-alive message
  if length == 0 {
    return nil, nil
  }

  messageBuf := make([]byte, length)
  _, err = io.ReadFull(r, messageBuf)
  if err != nil {
    return nil, err
  }

  m := Message{
    ID:       messageID(messageBuf[0]),
    Payload:  messageBuf[1:],
  }

  return &m, nil
}

func (m *Message) name() string {
  if m == nil {
    return "KeepAlive"
  }

  switch m.ID {
  case MsgChoke:
    return "Choke"
  case MsgUnchoke:
    return "Unchoke"
  case MsgInterested:
    return "Interested"
  case MsgNotInterested:
    return "MsgNotInterested"
  case MsgHave:
    return "Have"
  case MsgBitfield:
    return "Bitfield"
  case MsgRequest:
    return "Request"
  case MsgPiece:
    return "Piece"
  case MsgCancel:
    return "Cancel"
  default:
    return fmt.Sprintf("Unknown#%d", m.ID)
  }
}

func (m *Message) String() string {
	if m == nil {
		return m.name()
	}
	return fmt.Sprintf("%s [%d]", m.name(), len(m.Payload))
}
