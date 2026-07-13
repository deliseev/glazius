package service

import (
	"github.com/deliseev/glazius/internal/domain/entity"
)

func CalculateDiff(oldData, newData entity.TorrentData) []entity.TorrentFile {
	oldFilesMap := make(map[string]int, len(oldData.Files))
	for _, file := range oldData.Files {
		oldFilesMap[file.Path] = file.Size
	}

	var diff []entity.TorrentFile
	for _, newFile := range newData.Files {
		oldSize, exists := oldFilesMap[newFile.Path]

		if !exists || oldSize != newFile.Size {
			diff = append(diff, newFile)
		}
	}

	return diff
}
