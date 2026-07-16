package usecase

import (
	"context"
	"fmt"

	"github.com/deliseev/glazius/internal/domain/entity"
	"github.com/deliseev/glazius/internal/domain/port"
	"github.com/deliseev/glazius/internal/domain/service"
)

type CheckUpdatesUseCase struct {
	tracker port.TrackerClient
	repo    port.SeriesRepository
	storage port.TorrentStorage
	parser  port.TorrentParser
}

func NewCheckUpdatesUseCase(
	tracker port.TrackerClient,
	repo port.SeriesRepository,
	storage port.TorrentStorage,
	parser port.TorrentParser,
) *CheckUpdatesUseCase {
	return &CheckUpdatesUseCase{
		tracker: tracker,
		repo:    repo,
		storage: storage,
		parser:  parser,
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

	// Логика обновления
	torrentBytes, err := uc.tracker.DownloadTorrent(ctx, s.URL)
	if err != nil {
		return fmt.Errorf("failed to download torrent for %s: %w", s.Title, err)
	}

	newTorrentData, err := uc.parser.Parse(torrentBytes)
	if err != nil {
		return fmt.Errorf("failed to parse new torrent for %s: %w", s.Title, err)
	}

	// Получаем старые данные (если они есть)
	var oldTorrentData entity.TorrentData
	if s.LatestInfoHash != "" {
		oldData, err := uc.storage.Get(s.LatestInfoHash)
		if err != nil {
			return fmt.Errorf("failed to get stored torrent data for hash %s: %w", s.LatestInfoHash, err)
		}
		oldTorrentData, err = uc.parser.Parse(oldData)
		if err != nil {
			return fmt.Errorf("failed to parse torrent data: %w", err)
		}
	}

	diff := service.CalculateDiff(oldTorrentData, newTorrentData)
	if len(diff) > 0 {
		s.PendingAck = true
		s.LatestInfoHash = infoHash

		uc.storage.Save(infoHash, torrentBytes)

		err = uc.repo.Save(s)
		if err != nil {
			return fmt.Errorf("failed to save series %s: %w", s.Title, err)
		}
		fmt.Printf("Updated! Series: %s, Added files: %v\n", title, diff)
	}

	return nil
}
