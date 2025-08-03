package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)


func Test_HandshakeSerialize_OK(t *testing.T) {
	pstr := "BitTorrent protocol"
	infoHash := [20]byte{}
	peerID := [20]byte{}

	copy(infoHash[:], []byte("12345678901234567890"))
	copy(peerID[:], []byte("ABCDEFGHIJKLMNOPQRST"))

	hs := Handshake{
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerId:   peerID,
	}

	serialized := hs.Serialize()

	expectedLen := 49 + len(pstr)
	if len(serialized) != expectedLen {
		t.Fatalf("expected serialized length %d, got %d", expectedLen, len(serialized))
	}

	if serialized[0] != byte(len(pstr)) {
		t.Errorf("expected pstrlen %d, got %d", len(pstr), serialized[0])
	}

	if string(serialized[1:1+len(pstr)]) != pstr {
		t.Errorf("expected pstr %q, got %q", pstr, serialized[1:1+len(pstr)])
	}

	offset := 1 + len(pstr) + 8

	if !bytes.Equal(serialized[offset:offset+20], infoHash[:]) {
		t.Errorf("infoHash does not match")
	}

	if !bytes.Equal(serialized[offset+20:], peerID[:]) {
		t.Errorf("peerID does not match")
	}
}

func Test_HandshakeRead_OK(t *testing.T) {
	pstr := "BitTorrent protocol"
	pstrlen := byte(len(pstr))
	infoHash := [20]byte{}
	peerID := [20]byte{}
	copy(infoHash[:], []byte("12345678901234567890"))
	copy(peerID[:], []byte("ABCDEFGHIJKLMNOPQRST"))

	
	buf := make([]byte, 49+len(pstr))
	buf[0] = pstrlen
	copy(buf[1:], []byte(pstr))
	copy(buf[1+len(pstr)+8:], infoHash[:])
	copy(buf[1+len(pstr)+8+20:], peerID[:])

	handshake, err := ReadHandshake(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if handshake.Pstr != pstr {
		t.Errorf("expected Pstr %q, got %q", pstr, handshake.Pstr)
	}
	if !bytes.Equal(handshake.InfoHash[:], infoHash[:]) {
		t.Errorf("infoHash mismatch %q, got %q", infoHash[:], handshake.InfoHash[:])
	}
	if !bytes.Equal(handshake.PeerId[:], peerID[:]) {
		t.Errorf("peerID mismatch")
	}
}

func Test_HandshakeRead_Err(t *testing.T) {
	pstr := "BitTorrent protocol"
	pstrlen := byte(0) // len is zero
	infoHash := [20]byte{}
	peerID := [20]byte{}
	copy(infoHash[:], []byte("12345678901234567890"))
	copy(peerID[:], []byte("ABCDEFGHIJKLMNOPQRST"))

	
	buf := make([]byte, 49+len(pstr))
	buf[0] = pstrlen
	copy(buf[1:], []byte(pstr))
	copy(buf[1+len(pstr)+8:], infoHash[:])
	copy(buf[1+len(pstr)+8+20:], peerID[:])

	_, err := ReadHandshake(buf)
	
	require.Error(t, err, ErrPstrLenIsZero)
}

func Test_HandshakeSerializeThenRead_OK(t *testing.T) {
	pstr := "BitTorrent protocol"
	infoHash := [20]byte{}
	peerID := [20]byte{}
	copy(infoHash[:], []byte("12345678901234567890"))
	copy(peerID[:], []byte("ABCDEFGHIJKLMNOPQRST"))

	original := Handshake{
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerId:   peerID,
	}
	data := original.Serialize()

	parsed, err := ReadHandshake(data)
	if err != nil {
		t.Fatalf("ReadHandshake failed: %v", err)
	}
	if parsed.Pstr != original.Pstr {
		t.Errorf("Pstr mismatch")
	}
	if !bytes.Equal(parsed.InfoHash[:], original.InfoHash[:]) {
		t.Errorf("InfoHash mismatch")
	}
	if !bytes.Equal(parsed.PeerId[:], original.PeerId[:]) {
		t.Errorf("PeerID mismatch")
	}
}
