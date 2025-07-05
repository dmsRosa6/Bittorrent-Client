package bittorrent

type FileItem struct{
	Path string
	Size int32
	Offset int32
}

func (f *FileItem) New(path string, size int32, offset int32) FileItem {
	return FileItem{Path: path, Size: size, Offset: offset}
}

func (f *FileItem) GetPath() string {
	return f.Path
}

func (f *FileItem) GetSize() int32 {
	return f.Size
}

func (f *FileItem) GetOffset() int32 {
	return f.Offset
}