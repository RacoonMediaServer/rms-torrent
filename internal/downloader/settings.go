package downloader

type Settings struct {
	Input         []byte
	Destination   string
	DownloadLimit uint64
	UploadLimit   uint64
}
