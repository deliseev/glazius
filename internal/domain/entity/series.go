package entity

type Series struct {
	ID             string
	URL            string
	Title          string
	Description    string
	BaseInfoHash   string
	LatestInfoHash string
	TorrentLink    string
	PendingAck     bool
}
