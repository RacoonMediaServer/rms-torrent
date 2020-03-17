package rutor

import (
	"encoding/base64"
	"github.com/gocolly/colly/v2"
	"racoondev.tk/gitea/racoon/rtorrent/internal/types"
	"regexp"
	"strconv"
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
	torrent  types.Torrent

	peersParseExpr *regexp.Regexp
}

func newGrabber() *tableGrabber {
	return &tableGrabber{
		torrents: make([]types.Torrent, 0),
		state:    columnSkip,
		peersParseExpr: regexp.MustCompile(`\d+`),
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
		grabber.handlePeersColumn(e)
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

func (grabber *tableGrabber) handlePeersColumn(e *colly.HTMLElement) {
	peers := grabber.peersParseExpr.FindAllString(e.Text, -1)
	for _, peersCount := range peers {
		count, err := strconv.Atoi(peersCount)
		if err == nil {
			grabber.torrent.Peers += count
		}
	}


	grabber.torrents = append(grabber.torrents, grabber.torrent)
}
