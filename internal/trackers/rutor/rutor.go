package rutor

import (
	"encoding/base64"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/micro/go-micro/v2/logger"
	"github.com/parnurzeal/gorequest"
	"io"
	"net/url"
	"os"
	"racoondev.tk/gitea/racoon/rms-torrent/internal/types"
)

type SearchSession struct {
	c *colly.Collector
	r *gorequest.SuperAgent
}

const rutorDomain = "new-rutor.org"

func (session *SearchSession) Setup(settings types.SessionSettings) {
	session.c = colly.NewCollector(
		colly.UserAgent(settings.UserAgent),
		colly.AllowURLRevisit(),
	)

	session.r = gorequest.New()

	if settings.ProxyURL != "" {
		if err := session.c.SetProxy(settings.ProxyURL); err != nil {
			logger.Errorf("set proxy failed: %s", err.Error())
		}

		session.r = session.r.Proxy(settings.ProxyURL)
	}

	if settings.Debug {
		session.c.SetDebugger(&debug.LogDebugger{})
	}
}

func (session *SearchSession) SetCaptchaText(captchaText string) {

}

func (session *SearchSession) Search(text string) ([]types.Torrent, error) {

	grabber := newGrabber()

	session.c.OnHTML("#index > table > tbody > tr > td", func(e *colly.HTMLElement) {
		grabber.HandleColumn(e)
	})

	pageUrl := fmt.Sprintf("http://%s/search/0/1/000/0/%s", rutorDomain, url.QueryEscape(text))
	if err := session.c.Visit(pageUrl); err != nil {
		return nil, types.RaiseError(types.NetworkProblem, err)
	}

	session.c.Wait()

	return grabber.torrents, nil
}

func (session *SearchSession) Download(link, destination string) error {
	decoded, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		return types.RaiseError(types.NetworkProblem, err)
	}

	url := "http://" + rutorDomain + string(decoded)

	response, _, errors := session.r.Clone().Get(url).End()

	if errors != nil && len(errors) >= 0 {
		return types.RaiseError(types.NetworkProblem, errors[0])
	}
	defer response.Body.Close()

	file, err := os.Create(destination)
	if err != nil {
		return types.RaiseError(types.StorageProblem, err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	return types.RaiseError(types.StorageProblem, err)
}
