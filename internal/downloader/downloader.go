package downloader

type Downloader interface {
	Start()
	Files() []string
	Title() string
	Stop()
	Progress() float32
	IsComplete() bool
	SizeMB() uint64
	Close()
}
