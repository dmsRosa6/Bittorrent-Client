package bittorrent

import (
	"bytes"
	"testing"
)

func TestDecode(t *testing.T) {
	data := []byte(`d8:announce33:http://192.168.1.74:6969/announce7:comment17:Comment goes here10:created by25:Transmission/2.92 (14714)13:creation datei1460444420e8:encoding5:UTF-84:infod6:lengthi59616e4:name9:lorem.txt12:piece lengthi32768e6:pieces20:ABCDEFGHIJKLMNOPQRST7:privatei0eee`)
	bencoding := BEncoding{}
	
	torrent, err := bencoding.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	expected := &Torrent{
		Announce:     "http://192.168.1.74:6969/announce",
		Comment:      "Comment goes here",
		CreatedBy:    "Transmission/2.92 (14714)",
		CreationDate: 1460444420,
		Encoding:     "UTF-8",
		Name:         "lorem.txt",
		PieceSize:    32768,
		PieceHashes:  [][]byte{[]byte("ABCDEFGHIJKLMNOPQRST")},
		IsPrivate:    false,
		Files:        []FileItem{{Size: 59616, Path: "lorem.txt"}},
		BlockSize:    16 * 1024,
	}

	if torrent.Announce != expected.Announce {
		t.Errorf("announce: got %q, want %q", torrent.Announce, expected.Announce)
	}
	if torrent.Comment != expected.Comment {
		t.Errorf("comment: got %q, want %q", torrent.Comment, expected.Comment)
	}
	if torrent.CreatedBy != expected.CreatedBy {
		t.Errorf("createdBy: got %q, want %q", torrent.CreatedBy, expected.CreatedBy)
	}
	if torrent.CreationDate != expected.CreationDate {
		t.Errorf("creationDate: got %d, want %d", torrent.CreationDate, expected.CreationDate)
	}
	if torrent.Encoding != expected.Encoding {
		t.Errorf("encoding: got %q, want %q", torrent.Encoding, expected.Encoding)
	}
	if torrent.Name != expected.Name {
		t.Errorf("name: got %q, want %q", torrent.Name, expected.Name)
	}
	if torrent.PieceSize != expected.PieceSize {
		t.Errorf("pieceSize: got %d, want %d", torrent.PieceSize, expected.PieceSize)
	}
	if torrent.IsPrivate != expected.IsPrivate {
		t.Errorf("isPrivate: got %v, want %v", torrent.IsPrivate, expected.IsPrivate)
	}
	if torrent.BlockSize != expected.BlockSize {
		t.Errorf("blockSize: got %d, want %d", torrent.BlockSize, expected.BlockSize)
	}
	
	if len(torrent.PieceHashes) != len(expected.PieceHashes) {
		t.Errorf("pieceHashes length: got %d, want %d", len(torrent.PieceHashes), len(expected.PieceHashes))
	} else {
		for i, hash := range torrent.PieceHashes {
			if string(hash) != string(expected.PieceHashes[i]) {
				t.Errorf("pieceHash[%d]: got %x, want %x", i, hash, expected.PieceHashes[i])
			}
		}
	}
	
	if len(torrent.Files) != len(expected.Files) {
		t.Errorf("files length: got %d, want %d", len(torrent.Files), len(expected.Files))
	} else {
		for i := range torrent.Files {
			fileItem := &torrent.Files[i]
			if fileItem.Size != expected.Files[i].Size {
				t.Errorf("files[%d]. size: got %d, want %d", i, fileItem.Size, expected.Files[i].Size)
			}
			if fileItem.Path != expected.Files[i].Path {
				t.Errorf("files[%d]. path: got %q, want %q", i, fileItem.Path, expected.Files[i].Path)
			}
		}
	}
}

func TestEncode(t *testing.T){
	in := &Torrent{
		Announce:     "http://192.168.1.74:6969/announce",
		Comment:      "Comment goes here",
		CreatedBy:    "Transmission/2.92 (14714)",
		CreationDate: 1460444420,
		Encoding:     "UTF-8",
		Name:         "lorem.txt",
		PieceSize:    32768,
		PieceHashes:  [][]byte{[]byte("ABCDEFGHIJKLMNOPQRST")},
		IsPrivate:    false,
		Files:        []FileItem{{Size: 59616, Path: "lorem.txt"}},
		BlockSize:    16 * 1024,
	}

	expected := []byte(`d8:announce33:http://192.168.1.74:6969/announce7:comment17:Comment goes here10:created by25:Transmission/2.92 (14714)13:creation datei1460444420e8:encoding5:UTF-84:infod6:lengthi59616e4:name9:lorem.txt12:piece lengthi32768e6:pieces20:ABCDEFGHIJKLMNOPQRST7:privatei0eee`)
	
	bencoding := BEncoding{}

	raw, err := bencoding.Encode(*in)

	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(raw) != len(expected) {
		t.Errorf("Encode does not match. Different len. expected : %d, got : %d", len(expected), len(raw))
		return
	}

	for i, b := range(expected) {
		if b != raw[i] {
			t.Errorf("Encode does not match. expected: %s, got: %s", string(b), string(raw[i]))
		} 
	}
}

func TestDecodeThenEncode(t *testing.T) {
    originalData := []byte(`d8:announce33:http://192.168.1.74:6969/announce7:comment17:Comment goes here10:created by25:Transmission/2.92 (14714)13:creation datei1460444420e8:encoding5:UTF-84:infod6:lengthi59616e4:name9:lorem.txt12:piece lengthi32768e6:pieces20:ABCDEFGHIJKLMNOPQRST7:privatei0eee`)

    bencoding := BEncoding{}

    torrent, err := bencoding.Decode(originalData)
    if err != nil {
        t.Fatalf("Decode failed: %v", err)
    }

    encoded, err := bencoding.Encode(*torrent)
    if err != nil {
        t.Fatalf("Encode failed: %v", err)
    }

    if !bytes.Equal(encoded, originalData) {
        t.Errorf("Decode->Encode mismatch\nExpected: %s\nGot:      %s", string(originalData), string(encoded))
    }
}
