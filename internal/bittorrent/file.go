package bittorrent

type FileItem struct{
	Path string
	Size int
	Offset int
}

func (f *FileItem) New(path string, size int, offset int) FileItem {
	return FileItem{Path: path, Size: size, Offset: offset}
}

func (f *FileItem) GetPath() string {
	return f.Path
}

func (f *FileItem) GetSize() int {
	return f.Size
}

func (f *FileItem) GetOffset() int {
	return f.Offset
}

func (f *FileItem) FormattedSize() string {
	return bytesToString(f.Size)
}