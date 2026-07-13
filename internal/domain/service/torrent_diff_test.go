package service_test

import (
	"slices"
	"testing"

	"github.com/deliseev/glazius/internal/domain/entity"
	"github.com/deliseev/glazius/internal/domain/service"
)

func TestCalculateDiff(t *testing.T) {
	tests := []struct {
		name     string
		oldFiles []entity.TorrentFile
		newFiles []entity.TorrentFile
		wantDiff []entity.TorrentFile
	}{
		{
			name: "Нет изменений",
			oldFiles: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1000},
			},
			newFiles: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1000},
			},
			wantDiff: nil,
		},
		{
			name: "Добавлен один новый файл",
			oldFiles: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1000},
			},
			newFiles: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1000},
				{Path: "S01E02.mkv", Size: 1100},
			},
			wantDiff: []entity.TorrentFile{
				{Path: "S01E02.mkv", Size: 1100},
			},
		},
		{
			name: "Файл изменен (другой размер)",
			oldFiles: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1000},
			},
			newFiles: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1500}, // например, сделали repack или заменили на 1080p
			},
			wantDiff: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1500},
			},
		},
		{
			name: "Файл удален из раздачи (не должно быть в diff)",
			oldFiles: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1000},
				{Path: "Bonus.mp4", Size: 500},
			},
			newFiles: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1000},
			},
			wantDiff: nil,
		},
		{
			name:     "Старый торрент пустой (первое скачивание)",
			oldFiles: nil,
			newFiles: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1000},
			},
			wantDiff: []entity.TorrentFile{
				{Path: "S01E01.mkv", Size: 1000},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldData := entity.TorrentData{Files: tt.oldFiles}
			newData := entity.TorrentData{Files: tt.newFiles}

			gotDiff := service.CalculateDiff(oldData, newData)

			if !slices.Equal(gotDiff, tt.wantDiff) {
				t.Errorf("CalculateDiff() = %v, want %v", gotDiff, tt.wantDiff)
			}
		})
	}
}
