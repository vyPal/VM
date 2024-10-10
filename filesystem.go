package main

import (
	"io"
	"os"
	"path/filepath"
)

type VFS interface {
	Open(string) (interface{}, error)
	Close(interface{}) error
	Create(string) (interface{}, error)
	Remove(string) error
	Stat(string) (*FileInfo, error)
	ReadDir(string) ([]*FileInfo, error)
	Read(interface{}, []byte) (int, error)
	Write(interface{}, []byte) (int, error)
	ReadAt(interface{}, []byte, int64) (int, error)
	WriteAt(interface{}, []byte, int64) (int, error)
	Seek(interface{}, int64, int) (int64, error)
	LoadBinary(interface{}, *MemoryManager) uint32
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

func (f *FolderBasedFile) Close() error {
	return f.File.Close()
}

func (f *FolderBasedFile) Read(b []byte) (int, error) {
	return f.File.Read(b)
}

func (f *FolderBasedFile) Write(b []byte) (int, error) {
	return f.File.Write(b)
}

func (f *FolderBasedFile) ReadAt(b []byte, off int64) (int, error) {
	return f.File.ReadAt(b, off)
}

func (f *FolderBasedFile) WriteAt(b []byte, off int64) (int, error) {
	return f.File.WriteAt(b, off)
}

func (f *FolderBasedFile) Seek(off int64, whence int) (int64, error) {
	return f.File.Seek(off, whence)
}

func (vfs *FolderBasedVFS) Open(name string) (interface{}, error) {
	file, err := os.OpenFile(filepath.Join(vfs.Root, name), os.O_RDWR, 0644)
	return &FolderBasedFile{
		Name: name,
		File: file,
	}, err
}

func (vfs *FolderBasedVFS) Close(file interface{}) error {
	return file.(*FolderBasedFile).Close()
}

func (vfs *FolderBasedVFS) Create(name string) (interface{}, error) {
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

func (vfs *FolderBasedVFS) Read(file interface{}, b []byte) (int, error) {
	return file.(*FolderBasedFile).Read(b)
}

func (vfs *FolderBasedVFS) Write(file interface{}, b []byte) (int, error) {
	return file.(*FolderBasedFile).Write(b)
}

func (vfs *FolderBasedVFS) ReadAt(file interface{}, b []byte, off int64) (int, error) {
	return file.(*FolderBasedFile).ReadAt(b, off)
}

func (vfs *FolderBasedVFS) WriteAt(file interface{}, b []byte, off int64) (int, error) {
	return file.(*FolderBasedFile).WriteAt(b, off)
}

func (vfs *FolderBasedVFS) Seek(file interface{}, off int64, whence int) (int64, error) {
	return file.(*FolderBasedFile).Seek(off, whence)
}

func (vfs *FolderBasedVFS) LoadBinary(file interface{}, mm *MemoryManager) uint32 {
	f := file.(*FolderBasedFile)
	l, err := f.File.Seek(0, io.SeekEnd)
	if err != nil {
		panic(err)
	}
	f.File.Seek(0, io.SeekStart)
	data := make([]byte, l)
	_, err = f.File.Read(data)
	if err != nil {
		panic(err)
	}

	bc, err := DecodeBytecode(data)
	if err != nil {
		panic(err)
	}

	for _, sector := range bc.Sectors {
		if sector.Bytecode != nil {
			mm.Memory.LoadProgram(sector.StartAddress, sector.Bytecode)
		}
	}

	return bc.StartAddress
}