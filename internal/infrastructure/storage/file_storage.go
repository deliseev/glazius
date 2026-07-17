package storage

import (
	"os"
	"path/filepath"
)

type FileTorrentStorage struct {
	baseDir string
}

func NewFileTorrentStorage(dir string) (*FileTorrentStorage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FileTorrentStorage{baseDir: dir}, nil
}

func (s *FileTorrentStorage) Save(hash string, data []byte) error {
	return os.WriteFile(filepath.Join(s.baseDir, hash+".torrent"), data, 0644)
}

func (s *FileTorrentStorage) Get(hash string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.baseDir, hash+".torrent"))
}

func (s *FileTorrentStorage) CopyTo(hash, dest string) error {
	data, err := s.Get(hash)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0644)
}

func (s *FileTorrentStorage) Exists(infoHash string) bool {
	_, err := os.Stat(filepath.Join(s.baseDir, infoHash+".torrent"))
	if err == os.ErrExist {
		return true
	}
	return false
}
