package bittorrent

import (
	"crypto"
	"reflect"
)

type PieceVerifier struct {
	torrent	*Torrent
	fileManager	*FileManager
	VerifiedPieces chan<- int 
}

func NewPieceVerifier(torrent *Torrent, fm *FileManager) *PieceVerifier{
	return &PieceVerifier{
		torrent: torrent,
		fileManager: fm,
		VerifiedPieces: make(chan<- int, torrent.GetTotalPieces()),
	}
}

func (p *PieceVerifier) Verify(piece int) error {
	hash, err := p.getHash(piece)	

	if err != nil {
		return err 
	}

	isVerified := reflect.DeepEqual(hash, p.torrent.PieceHashes[piece])

	if isVerified {
		p.torrent.IsPieceVerified[piece] = true

		for i := 0; i < len(p.torrent.IsBlockAcquired[piece]);i++ {
			p.torrent.IsBlockAcquired[piece][i] = true
		}
	
		//TODO Check this later
		p.VerifiedPieces <- piece

		return nil
	}

	//if not verified something when wrong and lets reset it all
	p.torrent.IsPieceVerified[piece] = false
	if isArrayAllTrue(p.torrent.IsBlockAcquired[piece]) {
		setWholeArray(&p.torrent.IsBlockAcquired[piece],false)		
	}

	return nil 
}

func (pv *PieceVerifier) getHash(piece int) ([]byte, error) {
	raw, err := pv.fileManager.ReadPiece(piece)
	if err != nil {
		return nil, err
	}
	
	sha1 := crypto.SHA1.New()
	buf := sha1.Sum(raw) 
	return buf, nil 
}


func isArrayAllTrue(arr []bool) bool {

	for _,v := range(arr) {
		if !v {
			return false
		}
	}

	return true
}

func setWholeArray(arr *[]bool, val bool){

	for i := range(*arr) {
		(*arr)[i] = val
	}
}