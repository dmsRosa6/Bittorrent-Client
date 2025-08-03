package peer

import (
	"encoding/binary"
	"fmt"
)

type messageID byte

const (
    MsgChoke         messageID = 0
    MsgUnchoke       messageID = 1
    MsgInterested    messageID = 2
    MsgNotInterested messageID = 3
    MsgHave          messageID = 4
    MsgBitfield      messageID = 5
    MsgRequest       messageID = 6
    MsgPiece         messageID = 7
    MsgCancel        messageID = 8
)

// Message stores ID and payload of a message
type Message struct {
    ID      messageID
    Payload []byte
}

func (m *Message) Serialize() []byte {
    if m == nil {
        return make([]byte, 4)
    }
    length := uint32(len(m.Payload) + 1)
    buf := make([]byte, 4+length)
    binary.BigEndian.PutUint32(buf[0:4], length)
    buf[4] = byte(m.ID)
    copy(buf[5:], m.Payload)
    return buf
}

func ReadMessage(buf []byte) (*Message, error) {
    if len(buf) < 4 {
        return nil, fmt.Errorf("buffer too short to contain length: got %d", len(buf))
    }

    length := binary.BigEndian.Uint32(buf[0:4])
    
    if length == 0 {
        return nil, nil
    }

    if len(buf) < int(4+length) {
        return nil, fmt.Errorf("incomplete message: expected %d bytes, got %d", 4+length, len(buf))
    }

    id := messageID(buf[4])
    payload := make([]byte, length-1)
    copy(payload, buf[5:5+length-1])

    return &Message{
        ID:      id,
        Payload: payload,
    }, nil
}
