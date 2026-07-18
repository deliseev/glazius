package usecase

import (
	"context"
	"fmt"

	"github.com/deliseev/glazius/internal/domain/entity"
	"github.com/deliseev/glazius/internal/domain/port"
)

type CheckUpdatesUseCase struct {
	tracker port.TrackerClient
	repo    port.SeriesRepository
}

func NewCheckUpdatesUseCase(
	tracker port.TrackerClient,
	repo port.SeriesRepository,
) *CheckUpdatesUseCase {
	return &CheckUpdatesUseCase{
		tracker: tracker,
		repo:    repo,
	}
}

func (uc *CheckUpdatesUseCase) Execute(ctx context.Context) error {
	series, err := uc.repo.List()
	if err != nil {
		return fmt.Errorf("list series: %w", err)
	}

	for _, s := range series {
		if err := uc.checkSingleSeries(ctx, s); err != nil {
			fmt.Printf("error checking %s: %v\n", s.Title, err)
		}
	}
	return nil
}

func (uc *CheckUpdatesUseCase) checkSingleSeries(ctx context.Context, s entity.Series) error {
	title, infoHash, err := uc.tracker.FetchInfo(ctx, s.URL)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	if infoHash == s.LatestInfoHash {
		return nil
	}

	if s.LatestInfoHash != infoHash {
		s.PendingAck = true
		s.LatestInfoHash = infoHash

		err = uc.repo.Save(s)
		if err != nil {
			return fmt.Errorf("failed to save series %s: %w", s.Title, err)
		}
		fmt.Printf("Updated! Series: %s\n", title)
	}

	return nil
}
