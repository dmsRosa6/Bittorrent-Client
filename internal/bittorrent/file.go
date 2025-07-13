package bittorrent

type FileItem struct{
	Path string
	Size int64
	Offset int64
}

func (f *FileItem) New(path string, size int64, offset int64) FileItem {
	return FileItem{Path: path, Size: size, Offset: offset}
}

func (f *FileItem) GetPath() string {
	return f.Path
}

func (f *FileItem) GetSize() int64 {
	return f.Size
}

func (f *FileItem) GetOffset() int64 {
	return f.Offset
}

func (f *FileItem) FormattedSize() string {
	return bytesToString(f.Size)
}