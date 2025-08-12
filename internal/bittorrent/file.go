package bittorrent

import "sync"

type FileItem struct{
	Path string
	Size int
	Offset int
	mu sync.Mutex
}

func NewFileItem(path string, size int, offset int) FileItem {
	return FileItem{Path: path, Size: size, Offset: offset}
}
