package session

import (
	"encoding/json"
	"os"
	"time"

	bt "github.com/dmsosa6/bittorrent-client/internal/bittorrent"
)

// TODO have a way to save the session state so u can return to downloads
type Session struct {
	Torrents    map[bt.InfoHash]*bt.Torrent
	CurrTorrent *bt.Torrent
}

func NewSession() *Session {
	return &Session{
		Torrents: make(map[bt.InfoHash]*bt.Torrent),
	}
}

func (s *Session) AddTorrentToSession(t *bt.Torrent) {
	s.Torrents[t.InfoHash] = t
}

func (s *Session) SetCurrTorrent(t *bt.Torrent) {
	s.CurrTorrent = t
}

// TODO i will probabily remove this on the future
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

type PersistedTorrent struct {
	InfoHash    bt.InfoHash `json:"info_hash"`
	Downloaded  int64       `json:"downloaded"`
	TotalLength int64       `json:"total_length"`
	SavePath    string      `json:"save_path"`
	Bitfield    []bool      `json:"bitfield"`
}

type PersistedSession struct {
	Torrents    []PersistedTorrent `json:"torrents"`
	CurrTorrent bt.InfoHash        `json:"curr_torrent"`
}

func (s *Session) Save(path string) error {
	persisted := PersistedSession{}

	for _, t := range s.Torrents {
		pt := PersistedTorrent{
			InfoHash:    t.InfoHash,
			Downloaded:  t.BytesDownloaded,
			TotalLength: t.TotalLength,
			SavePath:    t.SavePath,
			Bitfield:    t.Bitfield,
		}
		persisted.Torrents = append(persisted.Torrents, pt)
	}

	if s.CurrTorrent != nil {
		persisted.CurrTorrent = s.CurrTorrent.InfoHash
	}

	data, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *Session) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var persisted PersistedSession
	if err := json.Unmarshal(data, &persisted); err != nil {
		return err
	}

	for _, pt := range persisted.Torrents {
		//need a way to reconstruct a Torrent
		t := bt.NewTorrentFromState(pt.InfoHash, pt.SavePath, pt.TotalLength, pt.Bitfield)
		s.Torrents[t.InfoHash] = t

		if pt.InfoHash == persisted.CurrTorrent {
			s.CurrTorrent = t
		}
	}

	return nil
}
