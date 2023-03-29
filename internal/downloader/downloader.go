package downloader

type Downloader interface {
	Start()
	Files() []string
	Title() string
	Stop()
	Progress() float32
	IsComplete() bool
	Close()
}

func New(settings Settings) (Downloader, error) {
	return newTorrentSession(settings)
}
