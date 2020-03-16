package types


type SearchSession interface {
	Setup(settings SessionSettings)
	Search(text string) ([]Torrent, error)
	Download(link, destination string) error
}

