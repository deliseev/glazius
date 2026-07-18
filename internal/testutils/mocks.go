package testutils

import (
	"context"

	"github.com/deliseev/glazius/internal/domain/entity"
)

type TrackerClientMock struct {
	FetchInfoFn func(ctx context.Context, url string) (string, string, string, error)
}

func (m *TrackerClientMock) FetchInfo(ctx context.Context, u string) (string, string, string, error) {
	return m.FetchInfoFn(ctx, u)
}

type RepoMock struct {
	ListFn func() ([]entity.Series, error)
	SaveFn func(s entity.Series) error
}

func (m *RepoMock) List() ([]entity.Series, error)       { return m.ListFn() }
func (m *RepoMock) Save(s entity.Series) error           { return m.SaveFn(s) }
func (m *RepoMock) Get(id string) (entity.Series, error) { return entity.Series{}, nil }
func (m *RepoMock) Delete(id string) error               { return nil }

type StorageMock struct {
	GetFn  func(hash string) ([]byte, error)
	SaveFn func(hash string, data []byte) error
}

func (m *StorageMock) Save(h string, d []byte) error { return m.SaveFn(h, d) }
func (m *StorageMock) Get(h string) ([]byte, error)  { return m.GetFn(h) }
func (m *StorageMock) CopyTo(h, p string) error      { return nil }
func (m *StorageMock) Exists(h string) bool          { return true }
