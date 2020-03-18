package types

type SessionSettings struct {
	User      string
	Password  string
	UserAgent string
	Debug     bool
}

type Torrent struct {
	Title        string
	DownloadLink string
	Size         string
	Peers        int
}
