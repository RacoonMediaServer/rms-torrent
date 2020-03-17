package rutracker

import (
	"encoding/base64"
	"github.com/gocolly/colly/v2"
	"racoondev.tk/gitea/racoon/rtorrent/internal/types"
)

type grabber struct {
	c       *colly.Collector
	torrent types.Torrent
	err     error
}

func spawnGrabber(c *colly.Collector, url string) *grabber {
	grabber := grabber{c: c}

	grabber.c.OnHTML("a.dl-link", func(e *colly.HTMLElement) {
		grabber.torrent.DownloadLink = base64.StdEncoding.EncodeToString([]byte(e.Attr("href")))
	})

	grabber.c.OnHTML("a.topic-title", func(e *colly.HTMLElement) {
		grabber.torrent.Title = e.Text
	})

	grabber.c.OnHTML(".borderless > b:nth-child(1)", func(e *colly.HTMLElement) {
		grabber.torrent.Size = e.Text
	})

	grabber.c.OnError(func(response *colly.Response, err error) {
		grabber.err = types.RaiseError(types.NetworkProblem, err)
	})

	grabber.err = types.RaiseError(types.NetworkProblem, grabber.c.Visit(url))
	return &grabber
}

func (grabber *grabber) Wait() (types.Torrent, error) {
	grabber.c.Wait()
	return grabber.torrent, grabber.err

}
