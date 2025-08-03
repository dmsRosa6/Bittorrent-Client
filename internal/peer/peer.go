package peer

import (
	"fmt"
	"net"
	"time"

	"github.com/dmsRosa6/bittorrent-client/internal/bittorrent"
)

const bufferSize = 256

type Peer struct{
	addr net.Addr
	LocalId string
	Id string
	torrent *bittorrent.Torrent
	IsDisconnected bool
	IsHandshakeSent bool
	IsPositionSent bool
	IsInterestedSent bool
	IsChokeSent bool
	IsHandShakeReceived bool
	IsChokeReceived bool
	IsInterestedReceived bool	
	IsBlockRequest [][]bool
	LastActive time.Time
	LastKeepActive time.Time
	Uploaded int64
	Download int64
}
//GO back to this
func CreatePeer(localId string, id string, isDisconencted bool, isHandshakeSent bool, isPosSent bool, isHandShakeReceived bool, lastActive time.Time, uploaded int64, downloaded int64) *Peer{
	return &Peer{
		LocalId: localId,
		Id: id,
		IsDisconnected: isDisconencted,
		IsHandshakeSent: isHandshakeSent,
		IsPositionSent: isPosSent,
		IsHandShakeReceived: isHandShakeReceived,
		LastActive: lastActive,
		Uploaded: uploaded,
		Download: downloaded,
	}
}

func (p * Peer) Connect() error{
	conn, err:= net.DialTimeout("tcp", "localhost:8080",3 * time.Second)
    if err != nil {
        fmt.Println("Error:", err)
        return err
    }
	buf := make([]byte, bufferSize)
	conn.Read(buf)

	sendHandShake()
	if(p.IsHandShakeReceived) {
		sendBitfield(p.torrent.IsPieceVerified)
	}

	return nil
}

func (p *Peer) Disconnected() error {
	return nil
}

func sendHandShake() error {
	return nil
}

func sendBitfield(IsPieceVerified []bool) error {

	return nil
}