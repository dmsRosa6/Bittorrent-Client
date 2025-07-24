package bittorrent

type PieceVerifier struct {
	torrent	*Torrent
	fileManager	*FileManager
}

func (p *PieceVerifier) Verify(piece int) error {

}

func (pv *PieceVerifier) getHash(piece int) ([]byte, error) {

}