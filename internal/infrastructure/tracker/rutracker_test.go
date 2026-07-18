package tracker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/deliseev/glazius/internal/domain/entity"
	"github.com/deliseev/glazius/internal/infrastructure/tracker"
)

func TestFetchInfo(t *testing.T) {
	tests := []struct {
		name      string
		fileName  string
		wantTitle string
		wantHash  string
		wantLink  string
		wantErr   bool
	}{
		{"Empty Page", "testdata/empty.html", "", "", "", true},
		{"Full Page", "testdata/sample.html", "Мыс страха", "50AD87D0A55D32440B91FA66EE24B71AD8A3E190", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlBody, err := os.ReadFile(tt.fileName)
			if err != nil {
				t.Fatalf("file %v not found", tt.fileName)
			}
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=windows-1251")
				w.Write(htmlBody)
			}))
			defer ts.Close()

			config := &entity.Config{}

			client, err := tracker.NewRutrackerClient(config)
			if err != nil {
				t.Error(err)
			}
			title, hash, link, err := client.FetchInfo(context.Background(), ts.URL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FetchInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !strings.Contains(title, tt.wantTitle) {
				t.Errorf("got %v, want %v", title, tt.wantTitle)
			}
			if hash != tt.wantHash {
				t.Errorf("got %v, want %v", hash, tt.wantHash)
			}
			if link != tt.wantLink {
				t.Errorf("got %v, want %v", link, tt.wantLink)
			}
		})
	}
}
