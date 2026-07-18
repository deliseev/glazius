package usecase_test

import (
	"context"
	"testing"

	"github.com/deliseev/glazius/internal/application/usecase"
	"github.com/deliseev/glazius/internal/domain/entity"
	"github.com/deliseev/glazius/internal/testutils"
)

func TestCheckUpdatesUseCase_Execute(t *testing.T) {
	oldHash := "old-hash"
	newHash := "new-hash"

	tracker := &testutils.TrackerClientMock{
		FetchInfoFn: func(ctx context.Context, url string) (string, string, string, error) {
			return "Title", newHash, "", nil
		},
		DownloadTorrentFn: func(ctx context.Context, url string) ([]byte, error) {
			return []byte("new-torrent-data"), nil
		},
	}

	repo := &testutils.RepoMock{
		ListFn: func() ([]entity.Series, error) {
			return []entity.Series{{ID: "1", URL: "url", LatestInfoHash: oldHash}}, nil
		},
		SaveFn: func(s entity.Series) error {
			if !s.PendingAck {
				t.Error("expected PendingAck to be true")
			}
			return nil
		},
	}

	storage := &testutils.StorageMock{
		GetFn:  func(hash string) ([]byte, error) { return []byte("old-data"), nil },
		SaveFn: func(h string, d []byte) error { return nil },
	}

	parser := &testutils.ParserMock{
		ParseFn: func(data []byte) (entity.TorrentData, error) {
			if string(data) == "new-torrent-data" {
				return entity.TorrentData{Files: []entity.TorrentFile{{Path: "new.mkv", Size: 100}}}, nil
			}
			return entity.TorrentData{Files: []entity.TorrentFile{{Path: "old.mkv", Size: 100}}}, nil
		},
	}

	uc := usecase.NewCheckUpdatesUseCase(tracker, repo, storage, parser)
	err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
