package gowebdav

import (
	"time"
)

// FileInfo 表示文件或目录的信息
type FileInfo interface {
	// Name 返回文件或目录的名称
	Name() string
	// Size 返回文件大小，目录通常返回0
	Size() int64
	// Mode 返回文件权限模式
	Mode() int
	// ModTime 返回最后修改时间
	ModTime() time.Time
	// IsDir 表示是否为目录
	IsDir() bool
}

// fileInfo 实现FileInfo接口的结构体
type fileInfo struct {
	name    string
	size    int64
	mode    int
	modTime time.Time
	isDir   bool
}

// NewFileInfo 创建一个新的FileInfo实例
func NewFileInfo(name string, size int64, mode int, modTime time.Time, isDir bool) FileInfo {
	return &fileInfo{
		name:    name,
		size:    size,
		mode:    mode,
		modTime: modTime,
		isDir:   isDir,
	}
}

// Name 返回文件或目录的名称
func (f *fileInfo) Name() string {
	return f.name
}

// Size 返回文件大小，目录通常返回0
func (f *fileInfo) Size() int64 {
	return f.size
}

// Mode 返回文件权限模式
func (f *fileInfo) Mode() int {
	return f.mode
}

// ModTime 返回最后修改时间
func (f *fileInfo) ModTime() time.Time {
	return f.modTime
}

// IsDir 表示是否为目录
func (f *fileInfo) IsDir() bool {
	return f.isDir
}
