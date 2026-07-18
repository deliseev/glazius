package usecase

import (
	"context"
	"fmt"

	"github.com/deliseev/glazius/internal/domain/entity"
	"github.com/deliseev/glazius/internal/domain/port"
	"github.com/google/uuid"
)

type AddSeriesUseCase struct {
	tracker    port.TrackerClient
	repository port.SeriesRepository
}

func NewAddSeriesUseCase(tracker port.TrackerClient, repository port.SeriesRepository) AddSeriesUseCase {
	return AddSeriesUseCase{
		tracker:    tracker,
		repository: repository,
	}
}

func (uc AddSeriesUseCase) Execute(ctx context.Context, url string) error {
	// 1. Получаем инфо (заголовок и хеш)
	title, infoHash, err := uc.tracker.FetchInfo(ctx, url)
	if err != nil {
		return fmt.Errorf("fetch error: %w", err)
	}

	// 4. Создаем сущность
	newSeries := entity.Series{
		ID:             uuid.NewString(),
		URL:            url,
		Title:          title,
		Description:    "",
		BaseInfoHash:   infoHash,
		LatestInfoHash: infoHash,
		PendingAck:     false,
	}
	err = uc.repository.Save(newSeries)
	if err != nil {
		return fmt.Errorf("save series error: %w", err)
	}

	return nil
}
