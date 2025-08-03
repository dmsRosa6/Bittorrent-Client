package tracker

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/dmsRosa6/bittorrent-client/internal/bittorrent"
)

//TODO Complete this test
func TestTracker(t *testing.T) {
	data := map[string]any{
	"announce": "http://192.168.1.74:6969/announce",
	"comment":  "Comment goes here",
	"created by": "Transmission/2.92 (14714)",
	"creation date": 1460444420,
	"encoding": "UTF-8",
	"info": map[string]any{
		"name":         "lorem.txt",
		"piece length": 32768, // 32 KiB
		"pieces":       string([]byte("ABCDEFGHIJKLMNOPQRSTABCDEFGHIJKLMNOPQRST")[:20]), // exactly 20 bytes
		"private":      0,
		"length":       59616, // single-file mode
	},
	}

	torrent, err := bittorrent.NewTorrent(data)

	if err != nil {
		log.Fatalf("failed to create torrent: %v", err)
	}


	trackerAddr := "localhost:6969"

    timeout := 2 * time.Second
    conn, err := net.DialTimeout("tcp", trackerAddr, timeout)
    if err != nil {
        t.Fatalf("OpenTracker is not running or unreachable at %s: %v", trackerAddr, err)
    }
    defer conn.Close()


	tracker := NewTracker(trackerAddr)
	err = tracker.Update(torrent,TrackerEvent(0),"1",8080)

}