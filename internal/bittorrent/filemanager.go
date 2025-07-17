package bittorrent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type FileManager struct {
	torrent Torrent
}

func (fm FileManager) Read(start int64, length int) ([]byte, error) {
	end := start + int64(length)
	buf := make([]byte, length)
	written := 0

	for _, file := range fm.torrent.Files {
		fileOffset := int64(file.Offset)
		fileSize := int64(file.Size)
		fileEnd := fileOffset + fileSize

		if end <= fileOffset || start >= fileEnd {
			continue
		}

		path := fm.torrent.DownloadDir + "/" + fm.torrent.FileDir() + file.Path
		exists, err := fileExists(path)
		
		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, fmt.Errorf("file does not exist. path: %s", path)
		}

		readStartInFile := maxInt64(0, start-fileOffset)
		readEndInFile := minInt64(fileSize, end-fileOffset)
		readLength := readEndInFile - readStartInFile

		writeOffset := written

		raw, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		defer raw.Close()

		_, err = raw.Seek(readStartInFile, 0)
		if err != nil {
			return nil, err
		}

		n, err := raw.Read(buf[writeOffset : writeOffset+int(readLength)])
		if err != nil && err.Error() != "EOF" {
			return nil, err
		}

		written += n
	}

	return buf, nil
}

func Write(start int64, buf []byte) error {


	return nil
} 
 

func fileExists(path string) (bool, error) {
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