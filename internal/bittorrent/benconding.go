package bittorrent

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

// BEncoding handles encoding and decoding of bencoded data.
// TODO Only the minimal subset required for .torrent files is implemented for now
type BEncoding struct{}

//This is how the structs are encoded and alligns with the protocol 
const (
    separator       = byte(':')
    integerStart    = byte('i')
    dictionaryStart = byte('d')
    listStart       = byte('l')
    objectEnd       = byte('e')
)

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

//form "<len>:<data>".
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

// TODO: convert the result into a proper Torrent struct once Torrent is defined.
func (b BEncoding) Decode(buf []byte) (*Torrent, error) {
    result, _, err := b.decodeAny(buf, 0)
    
    if err != nil {
        return nil, err
    }

    t, err := NewTorrent(result)
    
    if err != nil {
        return nil, err
    }

    return t, err
}

func (b BEncoding) DecodeFile(path string) (*Torrent, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    return b.Decode(data)
}

func (BEncoding) encodeNumber(val int) ([]byte,error) {
    return nil, nil
}

func (BEncoding) encodeString(val string) ([]byte,error) {
    return nil, nil
}

func (BEncoding) encodeList(val []interface{}) ([]byte,error) {
    return nil, nil
}

func (BEncoding) encodeMap(val map[string]interface{}) ([]byte,error) {
    return nil, nil
}


func (BEncoding) encodeAny(val map[string]interface{}) ([]byte,error) {
 // first verify if its a map
    return nil, nil
}

func (b BEncoding) Encode(t Torrent) ([]byte, error) {
    raw, err := t.ToBencodeMap()

    if err != nil {
        return nil, err
    }

    encoded, err := b.encodeAny(raw)

    if err != nil {
        return nil, err
    }

    return encoded, nil
}

func (b BEncoding) EncodeFile(name string, t Torrent) error {
    raw, err := b.Encode(t)

    if err != nil {
        return err
    }
    
    file, err := os.Create(name)

    if err != nil {
        return err
    }
    defer file.Close()

    // TODO here i can log how many files i wrote or smth
    _, err = file.Write(raw)

    if err != nil {
        return err
    }

    return nil
}
