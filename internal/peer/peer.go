package peer

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/dmsRosa6/bittorrent-client/internal/bittorrent"
)

const bufferSize = 4096 //  default buffer size for now

type Peer struct {
	Port     string
	Address  string
	Protocol string // tcp for BitTorrent (UDP is for tracker communication)
	LocalId  [20]byte
	Id       [20]byte
	torrent  *bittorrent.Torrent
	conn     net.Conn 
	
	// State flags
	IsDisconnected       bool
	IsHandshakeSent      bool
	IsHandshakeReceived  bool
	IsChokeReceived      bool
	IsInterestedReceived bool
	AmChoked            bool
	AmInterested        bool
	PeerChoked          bool
	PeerInterested      bool
	
	// Piece tracking
	HasPieces    []bool 
	IsBlockRequested [][]bool
	
	// Stats
	LastActive     time.Time
	LastKeepAlive  time.Time
	Uploaded       int64
	Downloaded     int64
}

func NewPeer(address, port string, torrent *bittorrent.Torrent, localId [20]byte) *Peer {
	numPieces := len(torrent.IsPieceVerified)
	return &Peer{
		Address:         address,
		Port:           port,
		Protocol:       "tcp",
		LocalId:        localId,
		torrent:        torrent,
		AmChoked:       true,  // Start choked
		PeerChoked:     true,
		HasPieces:      make([]bool, numPieces),
		IsBlockRequested: make([][]bool, numPieces),
		LastActive:     time.Now(),
	}
}

func (p *Peer) Connect() error {
	addr := net.JoinHostPort(p.Address, p.Port)
	conn, err := net.DialTimeout(p.Protocol, addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s:%s: %w", p.Address, p.Port, err)
	}
	
	p.conn = conn
	p.LastActive = time.Now()
	
	if err := p.sendHandshake(); err != nil {
		p.conn.Close()
		return fmt.Errorf("handshake failed: %w", err)
	}
	
	if err := p.readHandshakeResponse(); err != nil {
		p.conn.Close()
		return fmt.Errorf("failed to read handshake response: %w", err)
	}
	
	p.IsHandshakeSent = true
	p.IsHandshakeReceived = true
	
	return nil
}

func (p *Peer) Disconnect() error {
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	p.IsDisconnected = true
	return nil
}

func (p *Peer) sendHandshake() error {
	handshake := &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: p.torrent.InfoHash, 
		PeerId:   p.LocalId,
	}
	
	_, err := p.conn.Write(handshake.Serialize())
	return err
}

func (p *Peer) readHandshakeResponse() error {
	lengthBuf := make([]byte, 1)
	if _, err := p.conn.Read(lengthBuf); err != nil {
		return err
	}
	
	pstrLen := int(lengthBuf[0])
	if pstrLen == 0 {
		return fmt.Errorf("invalid pstr length: 0")
	}
	
	handshakeBuf := make([]byte, pstrLen+48)
	if _, err := p.conn.Read(handshakeBuf); err != nil {
		return err
	}
	
	fullHandshake := append(lengthBuf, handshakeBuf...)
	
	handshake, err := ReadHandshake(fullHandshake)
	if err != nil {
		return err
	}
	
	if handshake.InfoHash != p.torrent.InfoHash {
		return fmt.Errorf("info hash mismatch")
	}
	
	p.Id = handshake.PeerId
	
	return nil
}

func (p *Peer) SendBitfield() error {
	if p.conn == nil {
		return fmt.Errorf("not connected")
	}
	
	bitfield := p.createBitfield()
	message := p.createMessage(MsgBitfield, bitfield)
	
	_, err := p.conn.Write(message)
	return err
}

func (p *Peer) SendInterested() error {
	if p.conn == nil {
		return fmt.Errorf("not connected")
	}
	
	message := p.createMessage(MsgInterested, nil)
	_, err := p.conn.Write(message)
	if err == nil {
		p.AmInterested = true
	}
	return err
}

func (p *Peer) SendUnchoke() error {
	if p.conn == nil {
		return fmt.Errorf("not connected")
	}
	
	message := p.createMessage(MsgUnchoke, nil)
	_, err := p.conn.Write(message)
	if err == nil {
		p.PeerChoked = false
	}
	return err
}

func (p *Peer) SendKeepAlive() error {
	if p.conn == nil {
		return fmt.Errorf("not connected")
	}
	
	keepAlive := make([]byte, 4) // Length 0 message
	_, err := p.conn.Write(keepAlive)
	if err == nil {
		p.LastKeepAlive = time.Now()
	}
	return err
}

// Helper functions
func (p *Peer) createBitfield() []byte {
	numPieces := len(p.torrent.IsPieceVerified)
	bitfieldLen := (numPieces + 7) / 8 // Round up to nearest byte
	bitfield := make([]byte, bitfieldLen)
	
	for i, hasPiece := range p.torrent.IsPieceVerified {
		if hasPiece {
			byteIndex := i / 8
			bitIndex := uint(7 - (i % 8))
			bitfield[byteIndex] |= 1 << bitIndex
		}
	}
	
	return bitfield
}

func (p *Peer) createMessage(msgType byte, payload []byte) []byte {
	payloadLen := len(payload)
	messageLen := payloadLen + 1 // +1 for message type
	
	message := make([]byte, 4+messageLen) // 4 bytes for length prefix
	binary.BigEndian.PutUint32(message[0:4], uint32(messageLen))
	message[4] = msgType
	
	if payload != nil {
		copy(message[5:], payload)
	}
	
	return message
}

func (p *Peer) ReadMessage() (byte, []byte, error) {
	if p.conn == nil {
		return 0, nil, fmt.Errorf("not connected")
	}
	
	// Read message length
	lengthBuf := make([]byte, 4)
	if _, err := p.conn.Read(lengthBuf); err != nil {
		return 0, nil, err
	}
	
	messageLen := binary.BigEndian.Uint32(lengthBuf)
	
	// Keep-alive message
	if messageLen == 0 {
		p.LastActive = time.Now()
		return 0, nil, nil // Special case for keep-alive
	}
	
	// Read message
	messageBuf := make([]byte, messageLen)
	if _, err := p.conn.Read(messageBuf); err != nil {
		return 0, nil, err
	}
	
	p.LastActive = time.Now()
	
	msgType := messageBuf[0]
	payload := messageBuf[1:]
	
	return msgType, payload, nil
}

// Process incoming messages
func (p *Peer) HandleMessage(msgType byte, payload []byte) error {
	switch msgType {
	case MsgChoke:
		p.AmChoked = true
	case MsgUnchoke:
		p.AmChoked = false
	case MsgInterested:
		p.PeerInterested = true
	case MsgNotInterested:
		p.PeerInterested = false
	case MsgBitfield:
		return p.processBitfield(payload)
	case MsgHave:
		return p.processHave(payload)
	case MsgRequest:
		return p.processRequest(payload)
	case MsgPiece:
		return p.processPiece(payload)
	case MsgCancel:
		return p.processCancel(payload)
	}
	
	return nil
}

func (p *Peer) processBitfield(payload []byte) error {
	numPieces := len(p.torrent.IsPieceVerified)
	
	for i := 0; i < numPieces; i++ {
		byteIndex := i / 8
		bitIndex := uint(7 - (i % 8))
		
		if byteIndex < len(payload) {
			p.HasPieces[i] = (payload[byteIndex] & (1 << bitIndex)) != 0
		}
	}
	
	return nil
}

func (p *Peer) processHave(payload []byte) error {
	if len(payload) != 4 {
		return fmt.Errorf("invalid have message length")
	}
	
	pieceIndex := binary.BigEndian.Uint32(payload)
	if int(pieceIndex) < len(p.HasPieces) {
		p.HasPieces[pieceIndex] = true
	}
	
	return nil
}

func (p *Peer) processRequest(payload []byte) error {
	// TODO: Implement request handling
	return nil
}

func (p *Peer) processPiece(payload []byte) error {
	// TODO: Implement piece handling
	return nil
}

func (p *Peer) processCancel(payload []byte) error {
	// TODO: Implement cancel handling
	return nil
}

// Utility methods
func (p *Peer) IsConnected() bool {
	return p.conn != nil && !p.IsDisconnected
}

func (p *Peer) HasPiece(pieceIndex int) bool {
	if pieceIndex < 0 || pieceIndex >= len(p.HasPieces) {
		return false
	}
	return p.HasPieces[pieceIndex]
}