package client

import (
  "bytes"
  "net"
  "time"
  "fmt"

  "github.com/RRustom/go-bittorrent/bitfield"
  "github.com/RRustom/go-bittorrent/peers"
  "github.com/RRustom/go-bittorrent/handshake"
  "github.com/RRustom/go-bittorrent/message"
)

// a TCP connection with a peer
type Client struct {
  Conn      net.Conn
  Choked    bool
  Bitfield  bitfield.Bitfield
  peer      peers.Peer
  infoHash  [20]byte
  peerID    [20]byte
}

func completeHandshake(conn net.Conn, infohash, peerID [20]byte) (*handshake.Handshake, error) {
  conn.SetDeadline(time.Now().Add(3 * time.Second))
  defer conn.SetDeadline(time.Time{}) // disable the deadline

  req := handshake.New(infohash, peerID)
  _, err := conn.Write(req.Serialize())

  if err != nil {
    return nil, err
  }

  res, err := handshake.Read(conn)
  if err != nil {
    return nil, err
  }

  if !bytes.Equal(res.InfoHash[:], infohash[:]) {
    return nil, fmt.Errorf("Expected infohash %x, but got %x", res.InfoHash, infohash)
  }

  return res, nil
}

func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
  conn.SetDeadline(time.Now().Add(5 * time.Second))
  defer conn.SetDeadline(time.Time{}) // disable the deadline

  msg, err := message.Read(conn)

  if err != nil {
    return nil, err
  }
  if msg == nil {
    err := fmt.Errorf("Expected bitfield but got %s", msg)
    return nil, err
  }
  if msg.ID != message.MsgBitfield {
    err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
    return nil, err
  }

  return msg.Payload, nil
}

// connect with a peer, complete a handshake, receive a handshake
// returns an err if any of these fail
func New(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
  conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
  if err != nil {
    return nil, err
  }

  _, err = completeHandshake(conn, infoHash, peerID)
  if err != nil {
    conn.Close()
    return nil, err
  }

  bf, err := recvBitfield(conn)
  if err != nil {
    return nil, err
  }

  return &Client{
    Conn:     conn,
    Choked:   true,
    Bitfield: bf,
    peer:     peer,
    infoHash: infoHash,
    peerID:   peerID,
  }, nil
}

// read and consume a message from the connection
func (c *Client) Read() (*message.Message, error) {
  msg, err := message.Read(c.Conn)
  return msg, err
}

// send a request message to the peer
func (c *Client) SendRequest(index, begin, length int) error {
  req := message.FormatRequest(index, begin, length)
  _, err := c.Conn.Write(req.Serialize())
  return err
}

// send an Interested message to the peer
func (c *Client) SendInterested() error {
  msg := message.Message{ID: message.MsgInterested}
  _, err := c.Conn.Write(msg.Serialize())
  return err
}

// send a NotInterested message to the peer
func (c *Client) SendUnchoke() error {
  msg := message.Message{ID: message.MsgUnchoke}
  _, err := c.Conn.Write(msg.Serialize())
  return err
}

// send a Have message to the peer
func (c *Client) SendHave(index int) error {
  msg := message.FormatHave(index)
  _, err := c.Conn.Write(msg.Serialize())
  return err
}
