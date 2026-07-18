package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/deliseev/glazius/internal/application/usecase"
	"github.com/deliseev/glazius/internal/domain/entity"
	"github.com/deliseev/glazius/internal/infrastructure/storage"
)

type mockTrackerClient struct {
	title   string
	hash    string
	err     error
	torrent []byte
}

func (tc *mockTrackerClient) FetchInfo(ctx context.Context, url string) (title string, infoHash string, err error) {
	if tc.err != nil {
		return "", "", tc.err
	}
	return tc.title, tc.hash, nil
}

func (tc *mockTrackerClient) DownloadTorrent(ctx context.Context, url string) (torrentBytes []byte, err error) {
	if tc.err != nil {
		return []byte{}, err
	}
	return tc.torrent, nil
}

type mockRepository struct {
	store map[string]entity.Series
	err   error
}

func (r *mockRepository) Save(series entity.Series) error {
	if r.err != nil {
		return r.err
	}
	r.store[series.ID] = series
	return nil
}

func (r *mockRepository) List() ([]entity.Series, error) {
	if r.err != nil {
		return []entity.Series{}, r.err
	}
	result := make([]entity.Series, 0, len(r.store))
	for _, series := range r.store {
		result = append(result, series)
	}
	return result, nil
}

func (r *mockRepository) Get(id string) (entity.Series, error) {
	if r.err != nil {
		return entity.Series{}, r.err
	}
	return r.store[id], nil
}

func (r *mockRepository) Delete(id string) error {
	if r.err != nil {
		return r.err
	}
	delete(r.store, id)
	return nil
}

func TestAddSeriesUseCase(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
		count   int
	}{
		{"Success", nil, false, 1},
		{"TrackerError", fmt.Errorf("network error"), true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &mockTrackerClient{
				title:   "",
				hash:    "",
				err:     tt.mockErr,
				torrent: []byte{},
			}
			repository := &mockRepository{store: make(map[string]entity.Series)}
			storage := storage.NewMemoryStorage()
			useCase := usecase.NewAddSeriesUseCase(
				tracker,
				repository,
				storage,
			)
			ctx := context.Background()

			err := useCase.Execute(ctx, "test.url")

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr && err == nil {
				t.Fatalf("expected error %v, but empty", tt.mockErr)
			}
			if len(repository.store) != tt.count {
				t.Errorf("expected 1 series in store, got %d", len(repository.store))
			}
		})
	}
}
