package torrent_test

import (
	"os"
	"strings"
	"testing"

	"github.com/deliseev/glazius/internal/infrastructure/torrent"
)

func TestParse(t *testing.T) {
	// bencode данные: d4:infod5:filesld6:lengthi100e4:pathl5:file1.mkv...ee
	bencodeData, err := os.ReadFile("testdata/sample.torrent")
	if err != nil {
		t.Fatal("file sample.torrent not found")
	}

	parser := torrent.BencodeParser{}
	data, err := parser.Parse(bencodeData)

	if err != nil {
		t.Errorf("parse error: %v", err)
	}

	if len(data.Files) != 18 {
		t.Errorf("expected 18 files, got %d", len(data.Files))
	}
	if !strings.Contains(data.Files[0].Path, "S01") {
		t.Errorf("torrent file path not contained S01: %v", data.Files[0].Path)
	}
}
