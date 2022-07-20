package store

type Storage interface {
	UploadFile(path, name string, data []byte) error
	MakeDownloadURL(path, name string) string
	MakeRandomFilename() string
	DownloadFile(path, name string) ([]byte, error)
}
