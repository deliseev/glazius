package torrent

import (
	"bytes"
	"fmt"

	"path/filepath"

	"github.com/deliseev/glazius/internal/domain/entity"
	"github.com/zeebo/bencode"
)

type bencodeTorrent struct {
	Info struct {
		Files []struct {
			Length int      `bencode:"length"`
			Path   []string `bencode:"path"`
		} `bencode:"files"`
		Name string `bencode:"name"`
	} `bencode:"info"`
}

type BencodeParser struct {
}

func (p *BencodeParser) Parse(data []byte) (entity.TorrentData, error) {
	reader := bytes.NewReader(data)
	torrent := bencodeTorrent{}
	if err := bencode.NewDecoder(reader).Decode(&torrent); err != nil {
		return entity.TorrentData{}, fmt.Errorf("bencode decode: %w", err)
	}

	td := entity.TorrentData{}
	for _, file := range torrent.Info.Files {
		td.Files = append(td.Files, entity.TorrentFile{
			Path: filepath.Join(file.Path...),
			Size: file.Length,
		})
	}
	return td, nil
}
