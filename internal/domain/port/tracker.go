package port

import "context"

type TrackerClient interface {
	FetchInfo(ctx context.Context, url string) (title string, infoHash string, link string, err error)
}
