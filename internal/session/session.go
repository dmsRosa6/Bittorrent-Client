package session

import (
	"time"

	bt "github.com/dmsosa6/bittorrent-client/internal/bittorrent"
)

type Session struct {
	Torrents map[bt.InfoHash]*bt.Torrent
}

func NewSession() *Session {
	return &Session{
		Torrents: make(map[bt.InfoHash]*bt.Torrent),
	}
}

func (s *Session) AddTorrent(t *bt.Torrent) {
	s.Torrents[t.InfoHash] = t
}

// i will probabily remove this on the future
func (s *Session) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, torrent := range s.Torrents {
				torrent.UpdateTrackers()
				torrent.ManagePeers()
			}
		}
	}
}
