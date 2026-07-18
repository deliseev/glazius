package storage_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/deliseev/glazius/internal/infrastructure/storage"
)

func TestMemoryStorage_SaveAndGet(t *testing.T) {
	s := storage.NewMemoryStorage()
	hash := "test-hash-123"
	data := []byte("torrent-content-bytes")

	err := s.Save(hash, data)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	got, err := s.Get(hash)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if !bytes.Equal(got, data) {
		t.Errorf("got %s, want %s", got, data)
	}
}

func TestMemoryStorage_CopyTo(t *testing.T) {
	s := storage.NewMemoryStorage()
	hash := "test-hash-copy"
	data := []byte("content-to-copy")
	// Используем временную директорию ОС, чтобы не мусорить в папке проекта
	tempDir := t.TempDir()
	destFile := tempDir + "/test_out.torrent"

	s.Save(hash, data)

	err := s.CopyTo(hash, destFile)
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
