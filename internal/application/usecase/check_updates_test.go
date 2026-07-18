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
		FetchInfoFn: func(ctx context.Context, url string) (string, string, error) {
			return "Title", newHash, nil
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

	uc := usecase.NewCheckUpdatesUseCase(tracker, repo)
	err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
