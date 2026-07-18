package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/deliseev/glazius/internal/domain/entity"
)

type JSONRepository struct {
	path string
	mu   sync.RWMutex
	data map[string]entity.Series // ID -> Series
}

func NewJSONRepository(path string) (*JSONRepository, error) {
	repo := &JSONRepository{path: path, data: make(map[string]entity.Series)}
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return repo, nil // Файла нет — это ок, возвращаем пустой репозиторий
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(file, &repo.data)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *JSONRepository) Save(series entity.Series) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Не добавлять дубликаты сериалов
	for _, s := range r.data {
		if s.BaseInfoHash == series.BaseInfoHash {
			return fmt.Errorf("duplicate series: %v", series.BaseInfoHash)
		}
	}
	r.data[series.ID] = series

	file, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, file, 0644)
}

func (r *JSONRepository) List() ([]entity.Series, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var series []entity.Series
	for _, s := range r.data {
		series = append(series, s)
	}
	return series, nil
}

func (r *JSONRepository) Get(id string) (entity.Series, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if series, ok := r.data[id]; ok {
		return series, nil
	}
	return entity.Series{}, fmt.Errorf("not found series with id=%v", id)
}

func (r *JSONRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.data, id)

	file, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, file, 0644)
}
