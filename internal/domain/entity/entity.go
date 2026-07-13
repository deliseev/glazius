package entity

type TorrentFile struct {
	Path string
	Size int
}

type TorrentData struct {
	Files []TorrentFile
}
