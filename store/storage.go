package store

type Storage interface {
	UploadFile(path, name string, data []byte) error
	DownloadFile(path, name string) ([]byte, error)
}
