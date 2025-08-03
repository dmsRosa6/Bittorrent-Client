package tracker

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	bittorrent "github.com/dmsRosa6/bittorrent-client/internal/bittorrent"
)

type TrackerEvent int

const (
	StartedEvent TrackerEvent = iota
	PausedEvent
	StoppedEvent
)

var EventName = map[TrackerEvent]string{
	StartedEvent: "started",
	PausedEvent:  "paused",
	StoppedEvent: "stopped",
}

type Tracker struct {
	Address             string
	lastPeerRequest     time.Time
	peerRequestInterval time.Duration
	UpdatePeers		chan<- []net.Addr
}

func NewTracker(address string) *Tracker {
	c := make(chan []net.Addr)
	return &Tracker{
		Address:             address,
		lastPeerRequest:     time.Time{},
		peerRequestInterval: 30 * time.Minute,
		UpdatePeers:c,
	}
}

func (t *Tracker) Update(torrent *bittorrent.Torrent, ev TrackerEvent, id string, port int) error {
	now := time.Now().UTC()

	if ev == StartedEvent && now.Before(t.lastPeerRequest.Add(t.peerRequestInterval)) {
		return nil
	}

	t.lastPeerRequest = now

	url := fmt.Sprintf("%s?info_hash=%s&peer_id=%s&port=%d&uploaded=%d&downloaded=%d&left=%d&event=%s&compact=1",
		t.Address,
		torrent.UrlSafeStringInfohash(),
		id,
		port,
		torrent.Uploaded(),
		torrent.Downloaded(),
		torrent.Left(),
		EventName[ev],
	)

	t.request(url)
	return nil
}

func (t *Tracker) ResetLastRequest() {
	t.lastPeerRequest = time.Time{}
}

func (t *Tracker) request(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return err
	}

	bencode := bittorrent.BEncoding{}
	m, err := bencode.Decode(body)
	if err != nil {
		return err
	}

	trackerInfo := m.(map[string]any)
	t.peerRequestInterval, err = time.ParseDuration(trackerInfo["interval"].(string))
	if err != nil {
		return err
	}

	peerInfo := trackerInfo["peers"].([]byte)

	var peers []net.Addr

	for i := 0; i+6 <= len(peerInfo); i += 6 {
		ip := net.IPv4(peerInfo[i], peerInfo[i+1], peerInfo[i+2], peerInfo[i+3])
		port := int(binary.BigEndian.Uint16(peerInfo[i+4 : i+6]))
		addr := &net.TCPAddr{
			IP:   ip,
			Port: port,
		}
		peers = append(peers, addr)
	}

	t.UpdatePeers <- peers
	return nil
}