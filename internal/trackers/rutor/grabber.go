package rutor

import (
	"encoding/base64"
	"github.com/gocolly/colly/v2"
	"racoondev.tk/gitea/racoon/rtorrent/internal/types"
)

const (
	columnSkip = iota
	columnDate
	columnTitle
	columnSeparator
	columnSize
	columnPeers

	skipColumns = 4
)

type tableGrabber struct {
	torrents []types.Torrent
	state    int
	counter  int
	limit    uint
	torrent  types.Torrent
}

func newGrabber(limit uint) *tableGrabber {
	return &tableGrabber{
		torrents: make([]types.Torrent, 0),
		state:    columnSkip,
		limit:    limit,
	}
}

func (grabber *tableGrabber) HandleColumn(e *colly.HTMLElement) {
	grabber.counter++

	switch grabber.state {
	case columnSkip:
		if grabber.counter <= skipColumns {
			return
		}
		grabber.state = columnDate
		fallthrough

	case columnDate:
		grabber.torrent = types.Torrent{}

	case columnTitle:
		grabber.handleTitleColumn(e)

	case columnSize:
		grabber.torrent.Size = e.Text

	case columnPeers:
		if uint(len(grabber.torrents)) < grabber.limit {
			grabber.torrents = append(grabber.torrents, grabber.torrent)
		}

		grabber.state = columnDate
		return
	}

	grabber.state++
}

func (grabber *tableGrabber) handleTitleColumn(e *colly.HTMLElement) {
	grabber.torrent.Title = e.Text
	node := e.DOM.Find("a[href]")
	link, _ := node.Attr("href")
	grabber.torrent.DownloadLink = base64.StdEncoding.EncodeToString([]byte(link))
}
