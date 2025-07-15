package bittorrent

import (
	"errors"
	"fmt"
	"strings"
)

// TODO here is collapsed the info map that comes in the decode method, i hope this does not comes back to bite me, i dont see NOW why it would maybe the ordering matters. Because of this i need to build the info map on the encode
type Torrent struct {
	Announce       string      // URL of the tracker to connect to
	Name           string      // name of file or root directory
	IsPrivate      bool        // if true, peer exchange and DHT are disabled
	Files          []FileItem  // list of files
	AnnounceList   [][]string  // Additional trackers
	Comment        string      // human-readable comment
	Encoding       string      // what is used for encoding (top-level field)
	CreatedBy      string      // client that created the .torrent
	CreationDate   int       // when the .torrent was created
	DownloadDir    string      // local path where files will be saved
	BlockSize      int         // size of each block unit (e.g., 16 KiB)
	PieceSize      int         // size of each piece
	PieceHashes    [][]byte    // SHA-1 hashes for each piece
	IsPieceVerified []bool     // one per piece
	IsBlockAcquired [][]bool   // matrix [piece][block]
	downloaded     int       // total bytes downloaded
	uploaded       int       // total bytes uploaded
}

//TODO review this
// ToBencodeMap serializes the Torrent back into the nested map[string]any
// structure expected by your BEncoding.Encode method.
func (t Torrent) ToBencodeMap() (map[string]any, error) {
    top := make(map[string]any)

    // required
    top["announce"] = t.Announce

    // optional top‐level
    if len(t.AnnounceList) > 0 {
        // encode announce-list as []any of []any of string
        al := make([]any, len(t.AnnounceList))
        for i, tier := range t.AnnounceList {
            urls := make([]any, len(tier))
            for j, u := range tier {
                urls[j] = u
            }
            al[i] = urls
        }
        top["announce‐list"] = al
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

    // build info dict
    info := make(map[string]any, 6)

    // name is required
    info["name"] = t.Name
    // piece length required
    info["piece length"] = t.PieceSize
    // pieces: concat all SHA-1 hashes into one string
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
	}else{	
		info["private"] = 0
	}
	
    // files vs single‐file
    if len(t.Files) == 1 && t.Files[0].Path == t.Name {
        // single‐file mode: just length
        info["length"] = t.Files[0].Size
    } else {
        // multi‐file mode: list of dicts with length + path slice
        fl := make([]any, len(t.Files))
        for i, f := range t.Files {
            m := make(map[string]any, 2)
            m["length"] = f.Size
            // split the "/"‐joined Path back into components
            parts := strings.Split(f.Path, "/")
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

func NewTorrent(raw any) (*Torrent, error) {
	dic, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.New("top level must be a dictionary")
	}

	t := Torrent{
		BlockSize: 16 * 1024, // Default block size: 16 KiB
		Encoding:  "UTF-8",   // Default encoding
	}

	// Required top-level fields
	if announce, ok := dic["announce"].(string); ok {
		t.Announce = announce
	} else {
		return nil, errors.New("announce missing or not a string")
	}

	// Optional top-level fields
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

	// Info dictionary (required)
	infoDict, ok := dic["info"].(map[string]any)
	if !ok {
		return nil, errors.New("info dictionary missing")
	}

	// Required info fields
	if name, ok := infoDict["name"].(string); ok {
		t.Name = name
	} else {
		return nil, errors.New("name missing or not a string")
	}
	fmt.Print(infoDict["piece length"])
	if pieceLength, ok := infoDict["piece length"].(int); ok {
		t.PieceSize = int(pieceLength)
	} else {
		return nil, errors.New("piece length missing or not an integer")
	}

	// Handle pieces (required)
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

	// Optional info fields
	if private, ok := infoDict["private"].(int); ok {
		t.IsPrivate = private == 1
	}

	// Handle files (single/multi-file mode)
	if files, ok := infoDict["files"].([]any); ok {
		t.Files = make([]FileItem, 0, len(files))
		for _, f := range files {
			fileDict, ok := f.(map[string]any)
			if !ok {
				return nil, errors.New("file entry is not a dictionary")
			}
			length, ok := fileDict["length"].(int)
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
	} else if length, ok := infoDict["length"].(int); ok {
		t.Files = []FileItem{{
			Size: length,
			Path: t.Name,
		}}
	} else {
		return nil, errors.New("missing both 'files' and 'length' in info dict")
	}

	// Handle announce list (optional)
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

	totalSize := t.TotalSize()
	numPieces := len(t.PieceHashes)
	t.IsPieceVerified = make([]bool, numPieces)
	t.IsBlockAcquired = make([][]bool, numPieces)
	
	for i := 0; i < numPieces; i++ {
		pieceStart := int(i) * int(t.PieceSize)
		pieceEnd := pieceStart + int(t.PieceSize)
		if pieceEnd > totalSize {
			pieceEnd = totalSize
		}
		pieceSize := int(pieceEnd - pieceStart)

		numBlocks := pieceSize / t.BlockSize
		if pieceSize%t.BlockSize != 0 {
			numBlocks++
		}

		if numBlocks == 0 {
			numBlocks = 1
		}

		t.IsBlockAcquired[i] = make([]bool, numBlocks)
	}

	return &t, nil
}

func (t *Torrent) TotalSize() int {
	var size int
	for _, f := range t.Files {
		size += f.Size
	}
	return size
}

func (t *Torrent) FormattedPieceSize() string {
	return bytesToString(int(t.PieceSize))
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

func (t *Torrent) Uploaded() int {
	return t.uploaded
}

func (t *Torrent) Downloaded() int {
	return t.downloaded
}

func (t *Torrent) Left() int {
	return t.TotalSize() - t.downloaded
}

func bytesToString(val int) string {
	const unit = 1024
	if val < unit {
		return fmt.Sprintf("%d B", val)
	}
	div, exp := int(unit), 0
	for n := val / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(val)/float64(div), "KMGTPE"[exp])
}