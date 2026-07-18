package storage

import (
	"os"
)

type MemoryStorage struct {
	storage map[string][]byte
}

func (s *MemoryStorage) Save(infoHash string, data []byte) error {
	s.storage[infoHash] = data
	return nil
}

func (s *MemoryStorage) Get(infoHash string) ([]byte, error) {
	data, ok := s.storage[infoHash]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (s *MemoryStorage) CopyTo(infoHash string, destPath string) error {
	data, err := s.Get(infoHash)
	if err != nil {
		return err
	}
	// Реальная запись на диск, как и требует порт!
	return os.WriteFile(destPath, data, 0644)
}

func (s *MemoryStorage) Exists(infoHash string) bool {
	_, ok := s.storage[infoHash]
	return ok
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		storage: make(map[string][]byte),
	}
}
