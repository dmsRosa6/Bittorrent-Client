package bittorrent

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/dmsRosa6/bittorrent-client/internal/peer"
	"github.com/dmsRosa6/bittorrent-client/internal/tracker"
)

type InfoHash [20]byte

type PeerID [20]byte

type Torrent struct {
	// metadata
	Announce     string
	AnnounceList [][]string
	Comment      string
	CreatedBy    string
	CreationDate int
	Encoding     string

	Name        string
	IsPrivate   bool
	Files       []FileItem
	PieceSize   int
	PieceHashes [][]byte
	InfoHash    InfoHash
	InfoRaw     []byte

	// state
	DownloadDir     string
	BlockSize       int
	IsPieceVerified []byte
	IsBlockAcquired [][]bool
	OwnedPieces     []byte

	Downloaded int64
	Uploaded   int64

	// Swarm
	Peers    map[string]*peer.Peer
	Trackers []*tracker.Tracker

	IsPaused    bool
	IsSeeding   bool
	CompletedAt time.Time
}

func NewTorrent(dic map[string]any, rawInfoDict []byte) (*Torrent, error) {
	t := &Torrent{
		BlockSize: 16 * 1024,
		Encoding:  "UTF-8",
	}

	announce, ok := dic["announce"].(string)
	if !ok {
		return nil, errors.New("announce missing or not a string")
	}
	t.Announce = announce

	if encoding, ok := dic["encoding"].(string); ok {
		t.Encoding = encoding
	}
	if comment, ok := dic["comment"].(string); ok {
		t.Comment = comment
	}
	if createdBy, ok := dic["created by"].(string); ok {
		t.CreatedBy = createdBy
	}
	if creationDate, ok := dic["creation date"].(int); ok {
		t.CreationDate = creationDate
	}

	if rawInfoDict != nil {
		t.InfoHash = sha1.Sum(rawInfoDict)
	} else {
		return nil, errors.New("raw info dictionary required for InfoHash calculation")
	}

	infoDict, ok := dic["info"].(map[string]any)
	if !ok {
		return nil, errors.New("info dictionary missing")
	}

	name, ok := infoDict["name"].(string)
	if !ok {
		return nil, errors.New("name missing or not a string")
	}
	t.Name = name

	pieceLength, ok := infoDict["piece length"].(int)
	if !ok {
		return nil, errors.New("piece length missing or not an integer")
	}
	t.PieceSize = pieceLength

	pieces, ok := infoDict["pieces"].(string)
	if !ok {
		return nil, errors.New("pieces missing or not a string")
	}

	if len(pieces)%20 != 0 {
		return nil, errors.New("pieces string has invalid length")
	}

	numPieces := len(pieces) / 20
	t.PieceHashes = make([][]byte, numPieces)
	for i := 0; i < numPieces; i++ {
		hash := make([]byte, 20)
		copy(hash, pieces[i*20:(i+1)*20])
		t.PieceHashes[i] = hash
	}

	if private, ok := infoDict["private"].(int); ok {
		t.IsPrivate = private == 1
	}

	if err := t.parseFiles(infoDict); err != nil {
		return nil, err
	}

	if err := t.parseAnnounceList(dic); err != nil {
		return nil, err
	}

	t.initializeDownloadState()

	return t, nil
}

func (t *Torrent) parseFiles(infoDict map[string]any) error {
	if files, ok := infoDict["files"].([]any); ok {
		t.Files = make([]FileItem, 0, len(files))
		for _, f := range files {
			fileDict, ok := f.(map[string]any)
			if !ok {
				return errors.New("file entry is not a dictionary")
			}

			length, ok := fileDict["length"].(int)
			if !ok {
				return errors.New("file length missing or invalid")
			}

			pathList, ok := fileDict["path"].([]any)
			if !ok {
				return errors.New("file path missing or invalid")
			}

			pathComponents := make([]string, len(pathList))
			for i, comp := range pathList {
				if s, ok := comp.(string); ok {
					pathComponents[i] = s
				} else {
					return errors.New("path component not a string")
				}
			}

			t.Files = append(t.Files, FileItem{
				Size: length,
				Path: strings.Join(pathComponents, "/"),
			})
		}
	} else if length, ok := infoDict["length"].(int); ok {
		t.Files = []FileItem{{
			Size: length,
			Path: t.Name,
		}}
	} else {
		return errors.New("missing both 'files' and 'length' in info dict")
	}

	return nil
}

func (t *Torrent) parseAnnounceList(dic map[string]any) error {
	if announceList, ok := dic["announce-list"].([]any); ok {
		t.AnnounceList = make([][]string, 0, len(announceList))
		for _, tier := range announceList {
			if urls, ok := tier.([]any); ok {
				tierStrs := make([]string, 0, len(urls))
				for _, url := range urls {
					if s, ok := url.(string); ok {
						tierStrs = append(tierStrs, s)
					}
				}
				if len(tierStrs) > 0 {
					t.AnnounceList = append(t.AnnounceList, tierStrs)
				}
			}
		}
	}
	return nil
}

func (t *Torrent) initializeDownloadState() {
	numPieces := len(t.PieceHashes)

	t.IsPieceVerified = make([]bool, numPieces)
	t.IsBlockAcquired = make([][]bool, numPieces)

	for i := 0; i < numPieces; i++ {
		pieceSize := t.GetPieceSize(i)
		numBlocks := (pieceSize + t.BlockSize - 1) / t.BlockSize

		if numBlocks == 0 {
			numBlocks = 1
		}

		t.IsBlockAcquired[i] = make([]bool, numBlocks)
	}
}

func (t *Torrent) HexStringInfohash() string {
	return hex.EncodeToString(t.InfoHash[:])
}

func (t *Torrent) UrlSafeStringInfohash() string {
	return url.QueryEscape(string(t.InfoHash[:]))
}

func (t *Torrent) TotalSize() int {
	var size int
	for _, file := range t.Files {
		size += file.Size
	}
	return size
}

func (t *Torrent) PiecesCount() int {
	return len(t.PieceHashes)
}

func (t *Torrent) GetTotalPieces() int {
	return int(math.Ceil(float64(t.TotalSize()) / float64(t.PieceSize)))
}

func (t *Torrent) GetPieceSize(pieceIndex int) int {
	if pieceIndex < 0 || pieceIndex >= t.PiecesCount() {
		return 0
	}

	if pieceIndex == t.PiecesCount()-1 {
		remainder := t.TotalSize() % t.PieceSize
		if remainder != 0 {
			return remainder
		}
	}

	return t.PieceSize
}

func (t *Torrent) GetBlockSize(pieceIndex, blockIndex int) int {
	if pieceIndex < 0 || pieceIndex >= t.PiecesCount() {
		return 0
	}

	pieceSize := t.GetPieceSize(pieceIndex)
	numBlocks := (pieceSize + t.BlockSize - 1) / t.BlockSize

	if blockIndex < 0 || blockIndex >= numBlocks {
		return 0
	}

	if blockIndex == numBlocks-1 {
		remainder := pieceSize % t.BlockSize
		if remainder != 0 {
			return remainder
		}
	}

	return t.BlockSize
}

func (t *Torrent) Progress() float64 {
	if len(t.IsPieceVerified) == 0 {
		return 0.0
	}

	verified := 0
	for _, v := range t.IsPieceVerified {
		if v {
			verified++
		}
	}

	return float64(verified) / float64(len(t.IsPieceVerified))
}

func (t *Torrent) IsCompleted() bool {
	for _, verified := range t.IsPieceVerified {
		if !verified {
			return false
		}
	}
	return len(t.IsPieceVerified) > 0
}

func (t *Torrent) IsStarted() bool {
	return t.Downloaded > 0
}

func (t *Torrent) Left() int {
	remaining := t.TotalSize() - t.Downloaded
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (t *Torrent) FileDir() string {
	if len(t.Files) > 1 {
		return t.Name + "/"
	}
	return ""
}

func (t *Torrent) Validate() error {
	if t.Announce == "" {
		return errors.New("announce URL is required")
	}

	if t.Name == "" {
		return errors.New("torrent name is required")
	}

	if t.PieceSize <= 0 {
		return errors.New("piece size must be positive")
	}

	if len(t.PieceHashes) == 0 {
		return errors.New("torrent must have at least one piece")
	}

	if len(t.Files) == 0 {
		return errors.New("torrent must have at least one file")
	}

	for i, hash := range t.PieceHashes {
		if len(hash) != 20 {
			return fmt.Errorf("piece %d hash has invalid length %d, expected 20", i, len(hash))
		}
	}

	return nil
}

func (t *Torrent) IsMultiFile() bool {
	return len(t.Files) > 1
}

func (t *Torrent) GetAllTrackers() []string {
	trackers := []string{t.Announce}

	for _, tier := range t.AnnounceList {
		trackers = append(trackers, tier...)
	}

	return trackers
}

func (t *Torrent) FormattedPieceSize() string {
	return formatBytes(t.PieceSize)
}

func (t *Torrent) FormattedTotalSize() string {
	return formatBytes(t.TotalSize())
}

func formatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (t *Torrent) MarkPieceComplete(pieceIndex int) {
	if pieceIndex >= 0 && pieceIndex < len(t.IsPieceVerified) {
		t.IsPieceVerified[pieceIndex] = true

		if pieceIndex < len(t.IsBlockAcquired) {
			for i := range t.IsBlockAcquired[pieceIndex] {
				t.IsBlockAcquired[pieceIndex][i] = true
			}
		}
	}
}

func (t *Torrent) MarkBlockComplete(pieceIndex, blockIndex int) {
	if pieceIndex >= 0 && pieceIndex < len(t.IsBlockAcquired) {
		if blockIndex >= 0 && blockIndex < len(t.IsBlockAcquired[pieceIndex]) {
			t.IsBlockAcquired[pieceIndex][blockIndex] = true
		}
	}
}

func (t *Torrent) IsPieceComplete(pieceIndex int) bool {
	if pieceIndex < 0 || pieceIndex >= len(t.IsBlockAcquired) {
		return false
	}

	for _, hasBlock := range t.IsBlockAcquired[pieceIndex] {
		if !hasBlock {
			return false
		}
	}

	return true
}

func (t *Torrent) AddDownloaded(bytes int) {
	t.Downloaded += bytes
}

func (t *Torrent) AddUploaded(bytes int) {
	t.Uploaded += bytes
}

func (t *Torrent) ToBencodeMap() (map[string]any, error) {
	top := make(map[string]any)

	top["announce"] = t.Announce

	if len(t.AnnounceList) > 0 {
		al := make([]any, len(t.AnnounceList))
		for i, tier := range t.AnnounceList {
			urls := make([]any, len(tier))
			for j, u := range tier {
				urls[j] = u
			}
			al[i] = urls
		}
		top["announce-list"] = al
	}
	if t.Comment != "" {
		top["comment"] = t.Comment
	}
	if t.CreatedBy != "" {
		top["created by"] = t.CreatedBy
	}
	if t.CreationDate != 0 {
		top["creation date"] = t.CreationDate
	}
	if t.Encoding != "" {
		top["encoding"] = t.Encoding
	}

	info := make(map[string]any, 6)

	info["name"] = t.Name
	info["piece length"] = t.PieceSize
	concat := make([]byte, 0, len(t.PieceHashes)*20)
	for _, h := range t.PieceHashes {
		if len(h) != 20 {
			return nil, fmt.Errorf("invalid piece hash length %d, want 20", len(h))
		}
		concat = append(concat, h...)
	}
	info["pieces"] = string(concat)

	if t.IsPrivate {
		info["private"] = 1
	} else {
		info["private"] = 0
	}

	if len(t.Files) == 1 && t.Files[0].Path == t.Name {
		info["length"] = t.Files[0].Size
	} else {
		fl := make([]any, len(t.Files))
		for i := range t.Files {
			m := make(map[string]any, 2)
			m["length"] = t.Files[i].Size
			parts := strings.Split(t.Files[i].Path, "/")
			pa := make([]any, len(parts))
			for j, p := range parts {
				pa[j] = p
			}
			m["path"] = pa
			fl[i] = m
		}
		info["files"] = fl
	}

	top["info"] = info
	return top, nil
}

func (t *Torrent) ComputeInfoHash() [20]byte {
	return sha1.Sum(t.InfoRaw)
}

func (t *Torrent) InfoHashHex() string {
	h := t.ComputeInfoHash()
	return hex.EncodeToString(h[:])
}

func (t *Torrent) InfoHashURLEncoded() string {
	h := t.ComputeInfoHash()
	return url.QueryEscape(string(h[:]))
}

func (t *Torrent) RawInfo() []byte {
	return t.InfoRaw
}

func (t *Torrent) String() string {
	return "name blank for now. implement this"
}

func (t *Torrent) Details() string {
	return "details blank for now. implement this"
}
