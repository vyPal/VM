package main

import "os"

type VFS interface {
	Open(string) (*File, error)
	Close(*File) error
	Create(string) (*File, error)
	Remove(string) error
	Stat(string) (*FileInfo, error)
	ReadDir(string) ([]*FileInfo, error)
	Read(*File, []byte) (int, error)
	Write(*File, []byte) (int, error)
	ReadAt(*File, []byte, int64) (int, error)
	WriteAt(*File, []byte, int64) (int, error)
	Seek(*File, int64, int) (int64, error)
}

type File interface {
}

type FileInfo struct {
	Name string
	Size int64
	Mode os.FileMode
}

type FolderBasedVFS struct {
	Root string
}

type FolderBasedFile struct {
	Name string
	File *os.File
}

func (vfs *FolderBasedVFS) Open(name string) (*FolderBasedFile, error) {
	file, err := os.Open(name)
	return &FolderBasedFile{
		Name: name,
		File: file,
	}, err
}

func (vfs *FolderBasedVFS) Close(file *FolderBasedFile) error {
	return file.File.Close()
}

func (vfs *FolderBasedVFS) Create(name string) (*FolderBasedFile, error) {
	file, err := os.Create(name)
	return &FolderBasedFile{
		Name: name,
		File: file,
	}, err
}

func (vfs *FolderBasedVFS) Remove(name string) error {
	return os.Remove(name)
}

func (vfs *FolderBasedVFS) Stat(name string) (*FileInfo, error) {
	fileInfo, err := os.Stat(name)
	return &FileInfo{
		Name: name,
		Size: fileInfo.Size(),
		Mode: fileInfo.Mode(),
	}, err
}

func (vfs *FolderBasedVFS) ReadDir(name string) ([]*FileInfo, error) {
	files, err := os.ReadDir(name)
	fileInfos := make([]*FileInfo, 0, len(files))
	for _, file := range files {
		fileInfo, err := file.Info()
		if err != nil {
			return nil, err
		}
		fileInfos = append(fileInfos, &FileInfo{
			Name: fileInfo.Name(),
			Size: fileInfo.Size(),
			Mode: fileInfo.Mode(),
		})
	}
	return fileInfos, err
}

func (vfs *FolderBasedVFS) Read(file *FolderBasedFile, b []byte) (int, error) {
	return file.File.Read(b)
}

func (vfs *FolderBasedVFS) Write(file *FolderBasedFile, b []byte) (int, error) {
	return file.File.Write(b)
}

func (vfs *FolderBasedVFS) ReadAt(file *FolderBasedFile, b []byte, off int64) (int, error) {
	return file.File.ReadAt(b, off)
}

func (vfs *FolderBasedVFS) WriteAt(file *FolderBasedFile, b []byte, off int64) (int, error) {
	return file.File.WriteAt(b, off)
}

func (vfs *FolderBasedVFS) Seek(file *FolderBasedFile, off int64, whence int) (int64, error) {
	return file.File.Seek(off, whence)
}
