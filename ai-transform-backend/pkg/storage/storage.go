package storage

import "io"

type Storage interface {
	Upload(r io.Reader, md5Digest []byte, dstPath string) (url string, err error)
	UploadFromFile(path string, dstPath string) (url string, err error)
	DownloadFile(url string, dstPath string) error
}
type StorageFactory interface {
	CreateStorage() Storage
}
