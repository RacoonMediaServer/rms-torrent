package rutor

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"net/url"
	"racoondev.tk/gitea/racoon/rtorrent/internal/types"
)

type SearchSession struct {
	c *colly.Collector
}



func (session *SearchSession) Setup(settings types.SessionSettings) {
	session.c = colly.NewCollector(
		colly.UserAgent(settings.UserAgent),
		colly.Debugger(&debug.LogDebugger{}),
		colly.AllowURLRevisit(),
	)
}

func (session *SearchSession) SetCaptchaText(captchaText string) {

}

func (session *SearchSession) Search(text string, limit uint) ([]types.Torrent, error) {

	grabber := newGrabber(limit)

	session.c.OnHTML("#index > table > tbody > tr > td", func(e *colly.HTMLElement) {
		grabber.HandleColumn(e)
	})

	if err := session.c.Visit("http://new-rutor.org/search/0/1/000/0/" + url.QueryEscape(text)); err != nil {
		return nil, types.RaiseError(types.NetworkProblem, err)
	}

	session.c.Wait()

	return grabber.torrents, nil
}

func (session *SearchSession) Download(link, destination string) error {
	return nil
}
