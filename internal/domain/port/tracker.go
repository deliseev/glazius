package port

type TrackerClient interface {
	FetchInfo(url string) (title string, infoHash string, err error)
	DownloadTorrent(url string) (torrentBytes []byte, err error)
}
