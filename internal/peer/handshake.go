package peer

import (
	"errors"
	"fmt"
)


type Handshake struct {
	Pstr	string
	InfoHash	[20]byte
	PeerId	[20]byte
}

var ErrHandshakeInvalidLen = errors.New("handshake does not have the proper length")
var ErrPstrLenIsZero = errors.New("pstr length is 0")

func (h *Handshake) Serialize() []byte {
    buf := make([]byte, len(h.Pstr)+49)
    buf[0] = byte(len(h.Pstr))
    curr := 1
    curr += copy(buf[curr:], h.Pstr)
    curr += copy(buf[curr:], make([]byte, 8)) //reserved bytes
    curr += copy(buf[curr:], h.InfoHash[:])
    curr += copy(buf[curr:], h.PeerId[:])
    return buf
}

func Read(buf []byte) (*Handshake, error) {
    handshake := Handshake{}

    if len(buf) == 0 {
        return nil, fmt.Errorf("%w: got %d, expected at least %d", ErrHandshakeInvalidLen, len(buf), 49)
    }

    pstrLen := int(buf[0])

    if len(buf) < 49+pstrLen {
        return nil, fmt.Errorf("%w: got %d, expected %d", ErrHandshakeInvalidLen, len(buf), 49+pstrLen)
    }

    if pstrLen == 0 {
        return nil, ErrPstrLenIsZero
    }

    handshake.Pstr = string(buf[1 : 1+pstrLen])
    curr := 1 + pstrLen

    //reserved 8 bytes
    curr += 8

    copy(handshake.InfoHash[:], buf[curr:curr+20])
    curr += 20

    copy(handshake.PeerId[:], buf[curr:curr+20])
    curr += 20

    return &handshake, nil
}
