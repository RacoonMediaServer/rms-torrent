package downloader

type Downloader interface {
	Start()
	Files() []string
	Title() string
	Stop()
	Bytes() uint64
	RemainingBytes() uint64
	IsComplete() bool
	SizeMB() uint64
	Close()
}
