package repository_test

import (
	"path/filepath"
	"testing"

	"github.com/deliseev/glazius/internal/domain/entity"
	"github.com/deliseev/glazius/internal/infrastructure/repository"
)

type TestJSONRepository struct {
}

func TestJSONRepository_SaveAndList(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "data.json")

	repo, _ := repository.NewJSONRepository(filePath)
	series := entity.Series{ID: "1", Title: "Test"}

	// Тест сохранения
	repo.Save(series)

	// Тест чтения (пересоздаем репозиторий, чтобы прочитать с диска!)
	repo2, _ := repository.NewJSONRepository(filePath)
	list, _ := repo2.List()

	if len(list) != 1 || list[0].Title != "Test" {
		t.Errorf("failed to persist data to disk")
	}
}
