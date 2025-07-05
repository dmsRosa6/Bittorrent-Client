package bittorrent

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

// BEncoding handles encoding and decoding of bencoded data.
// TODO Only the minimal subset required for .torrent files is implemented for now.
type BEncoding struct{}

const (
    separator       = byte(':')
    integerStart    = byte('i')
    dictionaryStart = byte('d')
    listStart       = byte('l')
    objectEnd       = byte('e')
)

// decodeNumber parses an integer that starts after an 'i' and ends with 'e'.
// It returns the integer value, the next read position, and an error if any.
func (BEncoding) decodeNumber(buf []byte, pos int) (int, int, error) {
	if pos >= len(buf) {
		return 0, pos, io.ErrUnexpectedEOF
	}

	negative := false
	if buf[pos] == '-' {
		negative = true
		pos++
		if pos >= len(buf) {
			return 0, pos, errors.New("negative sign without number")
		}
	}

	if buf[pos] < '0' || buf[pos] > '9' {
		return 0, pos, errors.New("invalid digit in number")
	}

	if buf[pos] == '0' && pos+1 < len(buf) && buf[pos+1] != objectEnd {
		return 0, pos, errors.New("leading zeros are not allowed")
	}

	start := pos
	for pos < len(buf) && buf[pos] >= '0' && buf[pos] <= '9' {
		pos++
	}

	if pos >= len(buf) || buf[pos] != objectEnd {
		return 0, pos, errors.New("unterminated integer: missing 'e'")
	}

	numStr := string(buf[start:pos])
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, pos, fmt.Errorf("invalid number: %v", err)
	}

	if negative {
		if n == 0 {
			return 0, pos, errors.New("negative zero is not allowed")
		}
		n = -n
	}

	return n, pos + 1, nil
}


// decodeString parses a byte string of the form "<len>:<data>".
// It returns the decoded string, the next read position, and an error if any.
func (BEncoding) decodeString(buf []byte, pos int) (string, int, error) {
    start := pos
    for pos < len(buf) && buf[pos] != separator {
        if buf[pos] < '0' || buf[pos] > '9' {
            return "", pos, errors.New("invalid digit in string length")
        }
        pos++
    }
    if pos >= len(buf) {
        return "", pos, io.ErrUnexpectedEOF
    }

    length, err := strconv.Atoi(string(buf[start:pos]))
    if err != nil {
        return "", pos, err
    }

    pos++
    end := pos + length
    if end > len(buf) {
        return "", end, io.ErrUnexpectedEOF
    }

    return string(buf[pos:end]), end, nil
}

// decodeDictionary parses a bencoded dictionary.
func (b BEncoding) decodeDictionary(buf []byte, pos int) (map[string]any, int, error) {
    dict := make(map[string]any)
    head := pos
    for head < len(buf) && buf[head] != objectEnd {
        key, next, err := b.decodeString(buf, head)
        if err != nil {
            return nil, next, err
        }
        val, next2, err := b.decodeAny(buf, next)
        if err != nil {
            return nil, next2, err
        }
        dict[key] = val
        head = next2
    }
    if head >= len(buf) || buf[head] != objectEnd {
        return nil, head, io.ErrUnexpectedEOF
    }
    return dict, head + 1, nil
}

// decodeList parses a bencoded list.
func (b BEncoding) decodeList(buf []byte, pos int) ([]interface{}, int, error) {
    var list []interface{}
    head := pos
    for head < len(buf) && buf[head] != objectEnd {
        val, next, err := b.decodeAny(buf, head)
        if err != nil {
            return nil, next, err
        }
        list = append(list, val)
        head = next
    }
    if head >= len(buf) || buf[head] != objectEnd {
        return nil, head, io.ErrUnexpectedEOF
    }
    return list, head + 1, nil
}

// decodeAny dispatches decoding based on the leading byte.
func (b BEncoding) decodeAny(buf []byte, pos int) (interface{}, int, error) {
    if pos >= len(buf) {
        return nil, pos, io.ErrUnexpectedEOF
    }
    switch buf[pos] {
    case integerStart:
        return b.decodeNumber(buf, pos+1)
    case listStart:
        return b.decodeList(buf, pos+1)
    case dictionaryStart:
        return b.decodeDictionary(buf, pos+1)
    default:
        if buf[pos] >= '0' && buf[pos] <= '9' {
            return b.decodeString(buf, pos)
        }
    }
    return nil, pos, errors.New("unknown type prefix")
}

// Decode decodes a raw bencoded buffer.
// TODO: convert the result into a proper Torrent struct once Torrent is defined.
func (b BEncoding) Decode(buf []byte) (Torrent, error) {
    _, _, err := b.decodeAny(buf, 0)
    return Torrent{}, err
}

// DecodeFile reads a .torrent file and decodes its contents.
func (b BEncoding) DecodeFile(path string) (Torrent, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return Torrent{}, err
    }
    return b.Decode(data)
}

// Encode is not implemented yet.
func (BEncoding) Encode(_ Torrent) ([]byte, error) {
    return nil, errors.New("encode not implemented yet")
}

// EncodeFile is not implemented yet.
func (BEncoding) EncodeFile(_ string, _ Torrent) error {
    return errors.New("encode file not implemented yet")
}
