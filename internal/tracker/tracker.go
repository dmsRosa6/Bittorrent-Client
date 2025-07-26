package tracker

import (
	"fmt"
	"time"

	torrent "github.com/dmsRosa6/bittorrent-client/internal/bittorrent"
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
}

func NewTracker(address string) *Tracker {
	return &Tracker{
		Address:             address,
		lastPeerRequest:     time.Time{},              // zero time = no requests made yet
		peerRequestInterval: 30 * time.Minute,         // 30 minutes interval, can be configurable
	}
}

func (t *Tracker) Update(torrent *torrent.Torrent, ev TrackerEvent, id string, port int) {
	now := time.Now().UTC()

	// If event is Started and we haven't waited the full interval, skip request
	if ev == StartedEvent && now.Before(t.lastPeerRequest.Add(t.peerRequestInterval)) {
		return
	}

	// Update last request time
	t.lastPeerRequest = now

	// Build tracker announce URL with query parameters
	url := fmt.Sprintf("%s?info_hash=%s&peer_id=%s&port=%d&uploaded=%d&downloaded=%d&left=%d&event=%s&compact=1",
		t.Address,
		torrent.UrlSafeStringInfohash,
		id,
		port,
		torrent.Uploaded,
		torrent.Downloaded,
		torrent.Left,
		EventName[ev],
	)

	t.Request(url)
}

func (t *Tracker) ResetLastRequest() {
	t.lastPeerRequest = time.Time{}
}

func (t *Tracker) Request(url string) {
	// Your logic to send the request to the tracker goes here
	// e.g., an HTTP GET request
}

