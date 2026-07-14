package port

type TorrentStorage interface {
	Save(infoHash string, data []byte) error
	Get(infoHash string) ([]byte, error)
	CopyTo(infoHash string, destPath string) error
	Exists(infoHash string) bool
}
