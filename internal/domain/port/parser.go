package port

import "github.com/deliseev/glazius/internal/domain/entity"

type TorrentParser interface {
	Parse(data []byte) (entity.TorrentData, error)
}
