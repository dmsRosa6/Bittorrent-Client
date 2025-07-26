package bittorrent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)


type FileManager struct {
	torrent *Torrent
}

func (fm *FileManager) Read(start int64, length int) ([]byte, error) {
	end := start + int64(length)
	buf := make([]byte, length)
	written := 0

	for i := range fm.torrent.Files {
		fileItem := &fm.torrent.Files[i]
		fileOffset := int64(fileItem.Offset)
		fileSize := int64(fileItem.Size)
		fileEnd := fileOffset + fileSize

		if end <= fileOffset || start >= fileEnd {
			continue
		}

		path := fm.torrent.DownloadDir + "/" + fm.torrent.FileDir() + fileItem.Path
		exists, err := exists(path)
		
		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, fmt.Errorf("file does not exist. path: %s", path)
		}


		writeOffset := written

		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		readStartInFile := maxInt64(0, start-fileOffset)
		
		_, err = file.Seek(readStartInFile, 0)
		if err != nil {
			file.Close()
			return nil, err
		}
		
		readEndInFile := minInt64(fileSize, end-fileOffset)
		readLength := readEndInFile - readStartInFile

		n, err := file.Read(buf[writeOffset : writeOffset+int(readLength)])
		if err != nil && err.Error() != "EOF" {

			file.Close()
			return nil, err
		}

		written += n
		file.Close()
	}

	return buf, nil
}

func (fm *FileManager) Write(start int64, buf []byte) error {

	end := start + int64(len(buf))
	written := 0

	for i := range fm.torrent.Files {
		fileItem := &fm.torrent.Files[i]
		fileOffset := int64(fileItem.Offset)
		fileSize := int64(fileItem.Size)
		fileEnd := fileOffset + fileSize

		if end <= fileOffset || start >= fileEnd {
			continue
		}

		filePath := fm.torrent.DownloadDir + "/" + fm.torrent.FileDir() + fileItem.Path
		
		dirPath := fm.torrent.FileDir()

		exists, err := exists(dirPath)

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("folder does not exist. path: %s", dirPath)
		}

		
		file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		
		fileItem.mu.Lock()

		fileWriteStart := max(0, start - fileOffset)
		
		_, err = file.Seek(fileWriteStart, 0)
		if err != nil {
			fileItem.mu.Unlock()
			file.Close()
			return err
		}

		bufStart := max(0, fileOffset - start)
		bytesToWrite := min(fileEnd, end) - max(fileOffset, start)

		n, err := file.Write(buf[bufStart : bufStart+bytesToWrite])

		written += n

		if err != nil {
			fileItem.mu.Unlock()
			file.Close()
			return err 
		}

		fileItem.mu.Unlock()
		file.Close()
	}

	if written != len(buf) {
		return fmt.Errorf("buffer was only partially written starting at %d: expected %d bytes, wrote %d", start, len(buf), written)
 
	}

	return nil
} 

func (fm *FileManager) ReadPiece(piece int) ([]byte, error) {
	raw, err := fm.Read(int64(fm.torrent.PieceSize) * int64(piece), fm.torrent.GetPieceSize(piece))
	
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func (fm *FileManager) ReadBlock(piece int, offset int, length int) ([]byte, error) {
	raw, err := fm.Read(int64(fm.torrent.PieceSize) * int64(piece) + int64(offset), length)
	
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func (fm *FileManager) WriteBlock(piece int, block int, buf []byte) error {
	err := fm.Write(int64(fm.torrent.PieceSize) * int64(piece) + int64(block) * int64(fm.torrent.BlockSize), buf)
	
	if err != nil {
		return err
	}

	fm.torrent.IsBlockAcquired[piece][block] = true
	//TODO The hash needs to be verified, after the implementation
	//Verify(piece)
	return nil
}

func exists(path string) (bool, error) {
    _, err := os.Stat(path)

    if err == nil {
        return true, nil
    }
    if errors.Is(err, fs.ErrNotExist) {
        return false, nil
    }
    return false, err
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}