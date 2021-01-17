package torrentfile

import (
  "net/url"
  "net/http"
	"strconv"
  "time"
  "github.com/jackpal/bencode-go"
  "github.com/RRustom/go-bittorrent/peers"
)

type bencodeTrackerResponse struct {
  Interval  int     `bencode:"interval"`
  Peers     string  `bencode:"peers"`
}

func (*t TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
  base, err := url.Parse(t.Announce)
  if err != nil {
    return "", err
  }

  params := url.Values{
    "info_hash":  []string{string(t.InfoHash[:])},
    "peer_id":    []string{string(peerID[:])},
    "port":       []string{strconv.Itoa(int(Port))},
    "uploaded":   []string{"0"},
    "downloaded": []string{"0"},
    "compact":    []string{"1"},
    "left":       []string{strconv.Itoa(t.Length)},
  }

  base.RawQuery = params.Encode()
  return base.String(), nil
}

func (*t TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
  url, err := t.buildTrackerURL(peerID, port)
  if err != nil {
    return nil, err
  }

  c := &http.Client{Timeout: 15*time.Second}
  resp, err := c.Get(url)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  trackerResponse := bencodeTrackerResponse{}

  err = bencode.Unmarshal(resp.Body, &trackerResponse)
  if err != nil {
    return nil, err
  }

  return peers.Unmarshal([]byte(trackerResponse.Peers))
}