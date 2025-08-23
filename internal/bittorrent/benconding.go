package bittorrent

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// BEncoding handles encoding and decoding of bencoded data.
// TODO Only the minimal subset required for .torrent files is implemented for now
type BEncoding struct{}

// This is how the structs are encoded and alligns with the protocol
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

// form "<len>:<data>".
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

func (b BEncoding) decodeList(buf []byte, pos int) ([]any, int, error) {
	var list []any
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

func (b BEncoding) decodeAny(buf []byte, pos int) (any, int, error) {
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

func (b BEncoding) Decode(buf []byte) (any, error) {
	result, _, err := b.decodeAny(buf, 0)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b BEncoding) findRawInfo(buf []byte) ([]byte, error) {
	key := []byte("4:info")
	start := bytes.Index(buf, key)
	if start == -1 {
		return nil, fmt.Errorf("info dictionary not found")
	}

	start += len(key)

	_, end, err := b.decodeAny(buf, start)
	if err != nil {
		return nil, err
	}

	return buf[start:end], nil
}


// TODO i need to change this to have the raw info on a variable to create the torrent with it
func (b BEncoding) DecodeTorrent(buf []byte) (*Torrent, error) {
	result, _, err := b.decodeAny(buf, 0)

	if err != nil {
		return nil, err
	}

	infoRaw, err := b.findRawInfo(buf)
	if err != nil {
		return nil, err
	}

	t, err := NewTorrent(result.(map[string]any), infoRaw)

	if err != nil {
		return nil, err
	}

	return t, err
}

func (b BEncoding) DecodeTorrentFromFile(path string) (*Torrent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return b.DecodeTorrent(data)
}

// TODO Encode can have errors to so we need to check them
// POSITIVE NUMBERS ONLY AND ZEROS
// TYPES MISMATCH: INTERFACE CONVERTS ETC ...

func (BEncoding) encodeNumber(val int) []byte {

	s := string(integerStart) + strconv.Itoa(val) + string(objectEnd)
	buf := []byte(s)

	return buf
}

func (BEncoding) encodeString(val string) []byte {

	s := strconv.Itoa(len(val)) + string(separator) + val
	buf := []byte(s)

	return buf
}

func (b BEncoding) encodeList(l []any) []byte {

	buf := []byte{listStart}

	for _, item := range l {
		val := b.encodeAny(item)
		buf = append(buf, val...)
		buf = append(buf, separator)
	}

	buf[len(buf)-1] = objectEnd

	return buf
}

// Important reminder, the dic have lexicographic order
func (b BEncoding) encodeDictionary(m map[string]any) []byte {
	buf := []byte{dictionaryStart}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		key := b.encodeString(k)
		val := b.encodeAny(m[k])
		buf = append(buf, key...)
		buf = append(buf, val...)
	}

	buf = append(buf, objectEnd)
	return buf
}

func (b BEncoding) encodeAny(val any) []byte {

	switch v := val.(type) {

	case int:
		return b.encodeNumber(v)

	case string:
		return b.encodeString(v)

	case []any:
		return b.encodeList(v)

	case map[string]any:
		return b.encodeDictionary(v)

	default:
		return nil

	}
}

func (b BEncoding) Encode(t any) ([]byte, error) {
	return b.encodeAny(t), nil
}

func (b BEncoding) EncodeTorrent(t Torrent) ([]byte, error) {
	raw, err := t.ToBencodeMap()

	if err != nil {
		return nil, err
	}

	encoded := b.encodeAny(raw)

	return encoded, nil
}

func (b BEncoding) EncodeTorrentFromFile(name string, t Torrent) error {
	raw, err := b.EncodeTorrent(t)
	if err != nil {
		return err
	}

	err = safeWriteFile(name, raw)
	if err != nil {
		return err
	}

	return nil
}

func safeWriteFile(filename string, data []byte) error {
	dir := filepath.Dir(filename)
	tempFile, err := ioutil.TempFile(dir, "tmp-torrent-")
	if err != nil {
		return err
	}

	tempName := tempFile.Name()

	defer func() {
		tempFile.Close()
		os.Remove(tempName)
	}()

	_, err = tempFile.Write(data)
	if err != nil {
		return err
	}

	err = tempFile.Sync()
	if err != nil {
		return err
	}

	err = tempFile.Close()
	if err != nil {
		return err
	}

	return os.Rename(tempName, filename)
}
