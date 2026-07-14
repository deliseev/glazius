package storage_test

import (
	"bytes"
	"os"
	"testing"
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

// Здесь ты реализуешь свою структуру MemoryStorage,
// которая реализует интерфейс port.TorrentStorage

func TestMemoryStorage_SaveAndGet(t *testing.T) {
	storage := NewMemoryStorage() // Тебе нужно создать этот конструктор
	hash := "test-hash-123"
	data := []byte("torrent-content-bytes")

	err := storage.Save(hash, data)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	got, err := storage.Get(hash)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if !bytes.Equal(got, data) {
		t.Errorf("got %s, want %s", got, data)
	}
}

func TestMemoryStorage_CopyTo(t *testing.T) {
	storage := NewMemoryStorage()
	hash := "test-hash-copy"
	data := []byte("content-to-copy")
	// Используем временную директорию ОС, чтобы не мусорить в папке проекта
	tempDir := t.TempDir()
	destFile := tempDir + "/test_out.torrent"

	storage.Save(hash, data)

	err := storage.CopyTo(hash, destFile)
	if err != nil {
		t.Fatalf("CopyTo failed: %v", err)
	}

	// Проверяем, что файл реально создался и данные те же
	savedData, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !bytes.Equal(savedData, data) {
		t.Errorf("file content mismatch")
	}
}
