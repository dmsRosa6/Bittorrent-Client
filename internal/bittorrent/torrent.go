package bittorrent

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Torrent holds metadata and state for a BitTorrent download.
type Torrent struct {
	Announce       string      // URL of the tracker to connect to
	Name           string      // name of file or root directory
	IsPrivate      bool        // if true, peer exchange and DHT are disabled
	Files          []FileItem  // list of files
	AnnounceList   [][]string  // Additional trackers
	Comment        string      // human-readable comment
	CreatedBy      string      // client that created the .torrent
	CreationDate   time.Time   // when the .torrent was created
	DownloadDir    string      // local path where files will be saved
	BlockSize      int         // size of each block unit (e.g., 16 KiB)
	PieceSize      int         // size of each piece
	PieceHashes    [][]byte    // SHA-1 hashes for each piece
	IsPieceVerified []bool      // one per piece
	IsBlockAcquired [][]bool    // matrix [piece][block]
	downloaded     int64       // total bytes downloaded
	uploaded       int64       // total bytes uploaded
}

// TODO i dont initialize all fields
func NewTorrent(raw interface{}) (*Torrent, error) {
	dic, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.New("top level must be a dictionary")
	}

	t := Torrent{
		BlockSize: 16 * 1024, // Default block size: 16 KiB
	}

	if announce, ok := dic["announce"].(string); ok {
		t.Announce = announce
	} else {
		return nil, errors.New("announce missing or not a string")
	}

	infoDict, ok := dic["info"].(map[string]any)
	if !ok {
		return nil, errors.New("info dictionary missing")
	}

	if name, ok := infoDict["name"].(string); ok {
		t.Name = name
	} else {
		return nil, errors.New("name missing or not a string")
	}

	if pieceLength, ok := infoDict["piece length"].(int64); ok {
		t.PieceSize = int(pieceLength)
	} else {
		return nil, errors.New("piece length missing or not an integer")
	}

	if pieces, ok := infoDict["pieces"].(string); ok {
		if len(pieces)%20 != 0 {
			return nil, errors.New("pieces string has invalid length")
		}
		numPieces := len(pieces) / 20
		t.PieceHashes = make([][]byte, numPieces)
		for i := 0; i < numPieces; i++ {
			t.PieceHashes[i] = []byte(pieces[i*20 : (i+1)*20])
		}
	} else {
		return nil, errors.New("pieces missing or not a string")
	}

	if private, ok := infoDict["private"].(int64); ok {
		t.IsPrivate = private == 1
	}

	if files, ok := infoDict["files"].([]any); ok {
		t.Files = make([]FileItem, 0, len(files))
		for _, f := range files {
			fileDict, ok := f.(map[string]any)
			if !ok {
				return nil, errors.New("file entry is not a dictionary")
			}
			length, ok := fileDict["length"].(int64)
			if !ok {
				return nil, errors.New("file length missing or invalid")
			}
			pathList, ok := fileDict["path"].([]any)
			if !ok {
				return nil, errors.New("file path missing or invalid")
			}
			pathComponents := make([]string, len(pathList))
			for i, comp := range pathList {
				if s, ok := comp.(string); ok {
					pathComponents[i] = s
				} else {
					return nil, errors.New("path component not a string")
				}
			}
			t.Files = append(t.Files, FileItem{
				Size: length,
				Path: strings.Join(pathComponents, "/"),
			})
		}
	} else if length, ok := infoDict["length"].(int64); ok {
		t.Files = []FileItem{{
			Size: length,
			Path: t.Name,
		}}
	} else {
		return nil, errors.New("missing both 'files' and 'length' in info dict")
	}

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
				t.AnnounceList = append(t.AnnounceList, tierStrs)
			}
		}
	}

	if comment, ok := dic["comment"].(string); ok {
		t.Comment = comment
	}
	if createdBy, ok := dic["created by"].(string); ok {
		t.CreatedBy = createdBy
	}
	if creationDate, ok := dic["creation date"].(int64); ok {
		t.CreationDate = time.Unix(creationDate, 0)
	}

	totalSize := t.TotalSize()
	numPieces := len(t.PieceHashes)
	t.IsPieceVerified = make([]bool, numPieces)
	t.IsBlockAcquired = make([][]bool, numPieces)
	
	for i := 0; i < numPieces; i++ {
		pieceSize := t.PieceSize
		if i == numPieces-1 {
			lastPieceSize := int(totalSize) - i*t.PieceSize
			if lastPieceSize > 0 {
				pieceSize = lastPieceSize
			}
		}
		numBlocks := pieceSize / t.BlockSize
		if pieceSize%t.BlockSize != 0 {
			numBlocks++
		}
		t.IsBlockAcquired[i] = make([]bool, numBlocks)
	}

	return &t, nil
}

func (t *Torrent) TotalSize() int64 {
	var size int64
	for _, f := range t.Files {
		size += f.Size
	}
	return size
}

func (t *Torrent) FormattedPieceSize() string {
	return bytesToString(int64(t.PieceSize))
}

func (t *Torrent) FormattedTotalSize() string {
	return bytesToString(t.TotalSize())
}

func (t *Torrent) PiecesCount() int {
	return len(t.PieceHashes)
}

func (t *Torrent) IsStarted() bool {
	return t.downloaded > 0
}

func (t *Torrent) IsCompleted() bool {
	for _, verified := range t.IsPieceVerified {
		if !verified {
			return false
		}
	}
	return true
}

func (t *Torrent) Uploaded() int64 {
	return t.uploaded
}

func (t *Torrent) Downloaded() int64 {
	return t.downloaded
}

func (t *Torrent) Left() int64 {
	return t.TotalSize() - t.downloaded
}

func bytesToString(val int64) string {
	const unit = 1024
	if val < unit {
		return fmt.Sprintf("%d B", val)
	}
	div, exp := int64(unit), 0
	for n := val / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(val)/float64(div), "KMGTPE"[exp])
}