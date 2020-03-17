package types

type CaptchaHandler func(Url string) (string, error)

type SessionSettings struct {
	User      string
	Password  string
	UserAgent string
}

type Torrent struct {
	Title        string
	DownloadLink string
	Size         string
	Peers        int
}
