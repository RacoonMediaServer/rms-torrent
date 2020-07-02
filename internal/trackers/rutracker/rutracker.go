package rutracker

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
	"regexp"
	"strconv"
)

type captchaInfo struct {
	IsPresent  bool
	Url        string
	Sid        string
	Code       string
	Recognized string
}

type SearchSession struct {
	c          *colly.Collector
	r *gorequest.SuperAgent
	settings   types.SessionSettings
	authorized bool

	captchaSidExpr  *regexp.Regexp
	captchaCodeExpr *regexp.Regexp
	captchaUrlExpr  *regexp.Regexp

	captcha captchaInfo
}

func (session *SearchSession) Setup(settings types.SessionSettings) {
	session.settings = settings
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

	session.captchaSidExpr = regexp.MustCompile(`<input[^>]*name="cap_sid"[^>]*value="([^"]+)"[^>]*>`)
	session.captchaCodeExpr = regexp.MustCompile(`<input[^>]*name="(cap_code_[^"]+)"[^>]*value="[^"]*"[^>]*>`)
	session.captchaUrlExpr = regexp.MustCompile(`<img[^>]*src="([^"]+\/captcha\/[^"]+)"[^>]*>`)

	session.captcha = captchaInfo{}
}

func (session *SearchSession) SetCaptchaText(captchaText string) {
	session.captcha.Recognized = captchaText
}

func (session *SearchSession) Search(text string) ([]types.Torrent, error) {
	if err := session.authorize(); err != nil {
		return nil, err
	}

	torrents := make([]types.Torrent, 0)

	session.c.OnHTML("#tor-tbl > tbody > tr", func(e *colly.HTMLElement) {
		torrents = append(torrents, extractTorrent(e))
	})

	if err := session.c.Visit("https://rutracker.org/forum/tracker.php?nm=" + url.QueryEscape(text)); err != nil {
		return nil, types.RaiseError(types.NetworkProblem, err)
	}

	session.c.Wait()

	if !session.authorized && session.captcha.IsPresent {
		return nil, types.Error{
			Underlying: nil,
			Code:       types.CaptchaRequired,
			Captcha:    session.captcha.Url,
		}
	}

	if len(torrents) == 0 && !session.authorized {
		return nil, types.Error{Code: types.AuthFailed}
	}

	return torrents, nil
}

func (session *SearchSession) Download(link, destination string) error {
	cookies := session.c.Cookies("https://rutracker.org/forum/tracker.php")

	decoded, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		return types.RaiseError(types.NetworkProblem, err)
	}

	url := "https://rutracker.org/forum/" + string(decoded)
	request := session.r.Clone()
	for _, cookie := range cookies {
		request = request.AddCookie(cookie)
	}

	response, _, errors := request.Get(url).End()
	if errors != nil {
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

func (session *SearchSession) authorize() error {
	if !session.authorized {
		session.c.OnHTML("#logged-in-username", func(e *colly.HTMLElement) {
			session.authorized = true
			session.captcha.IsPresent = false
		})

		session.c.OnResponse(func(response *colly.Response) {
			session.extractCaptcha(response.Body)
		})

		var err error
		if session.captcha.IsPresent {
			err = session.c.Post("https://rutracker.org/forum/login.php", map[string]string{
				"login_username":     session.settings.User,
				"login_password":     session.settings.Password,
				"login":              "вход",
				"cap_sid":            session.captcha.Sid,
				session.captcha.Code: session.captcha.Recognized,
			})
		} else {
			err = session.c.Post("https://rutracker.org/forum/login.php", map[string]string{
				"login_username": session.settings.User,
				"login_password": session.settings.Password,
				"login":          "вход",
			})
		}

		return types.RaiseError(types.AuthFailed, err)
	}

	return nil
}

func (session *SearchSession) extractCaptcha(data []byte) {
	content := string(data)

	matches := session.captchaUrlExpr.FindStringSubmatch(content)
	if len(matches) < 2 {
		session.captcha.IsPresent = false
		return
	}

	session.captcha.IsPresent = true
	session.authorized = false
	session.captcha.Url = matches[1]

	matches = session.captchaCodeExpr.FindStringSubmatch(content)
	if len(matches) >= 2 {
		session.captcha.Code = matches[1]
	}

	matches = session.captchaSidExpr.FindStringSubmatch(content)
	if len(matches) >= 2 {
		session.captcha.Sid = matches[1]
	}

	fmt.Printf("%+v\n", session.captcha)
}

func extractTorrent(e *colly.HTMLElement) types.Torrent {
	torrent := types.Torrent{}
	torrent.Title = e.DOM.Find(`a.tLink`).Text()

	dl := e.DOM.Find(`a.tr-dl`)
	link, _ := dl.Attr("href")
	torrent.DownloadLink = base64.StdEncoding.EncodeToString([]byte(link))
	torrent.Size = dl.Text()

	seeds := e.DOM.Find(`b.seedmed`).Text()
	torrent.Peers, _ = strconv.Atoi(seeds)

	leechs := e.DOM.Find(`td.leechmed`).Text()
	peers, _ := strconv.Atoi(leechs)
	torrent.Peers += peers

	return torrent
}
