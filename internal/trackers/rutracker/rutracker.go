package rutracker

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"net/url"
	"racoondev.tk/gitea/racoon/rtorrent/internal/types"
	"racoondev.tk/gitea/racoon/rtorrent/internal/utils"
	"regexp"
)

type captchaInfo struct {
	IsPresent bool
	Url       string
	Sid       string
	Code      string
	Decoded   string
}

type SearchSession struct {
	c          *colly.Collector
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
		colly.Debugger(&debug.LogDebugger{}),
		colly.AllowURLRevisit(),
	)

	session.captchaSidExpr = regexp.MustCompile(`<input[^>]*name="cap_sid"[^>]*value="([^"]+)"[^>]*>`)
	session.captchaCodeExpr = regexp.MustCompile(`<input[^>]*name="(cap_code_[^"]+)"[^>]*value="[^"]*"[^>]*>`)
	session.captchaUrlExpr = regexp.MustCompile(`<img[^>]*src="([^"]+\/captcha\/[^"]+)"[^>]*>`)

	session.captcha = captchaInfo{}
}

func (session *SearchSession) Search(text, captchaText string) ([]types.Torrent, error) {
	session.captcha.Decoded = captchaText

	if err := session.authorize(); err != nil {
		return nil, err
	}

	children := make([]*grabber, 0)
	torrents := make([]types.Torrent, 0)

	session.c.OnHTML("a.tLink", func(e *colly.HTMLElement) {
		children = append(children, spawnGrabber(session.c.Clone(), "https://rutracker.org/forum/"+e.Attr("href")))
	})

	if err := session.c.Visit("https://rutracker.org/forum/tracker.php?nm=" + url.QueryEscape(text)); err != nil {
		return nil, types.RaiseError(types.NetworkProblem, err)
	}

	session.c.Wait()
	var lastError error
	for _, child := range children {
		torrent, err := child.Wait()
		if err == nil {
			torrents = append(torrents, torrent)
		} else {
			lastError = err
		}
	}

	if len(torrents) == 0 && session.captcha.IsPresent {
		return nil, types.Error{
			Underlying: nil,
			Code:       types.CaptchaRequired,
			Captcha:    session.captcha.Url,
		}
	}

	if len(torrents) == 0 && lastError != nil {
		return nil, lastError
	}

	if len(torrents) == 0 && !session.authorized {
		return nil, types.Error{Code: types.AuthFailed}
	}

	return torrents, nil
}

func (session *SearchSession) Download(link, destination string) error {
	return nil
}

func (session *SearchSession) authorize() error {
	if !session.authorized {
		session.c.OnHTML("#logged-in-username", func(e *colly.HTMLElement) {
			session.authorized = true
			session.captcha.IsPresent = false
		})

		session.c.OnResponse(func(response *colly.Response) {
			utils.DumpPage("auth", response.Body)
			session.extractCaptcha(response.Body)
		})

		var err error
		if session.captcha.IsPresent {
			err = session.c.Post("https://rutracker.org/forum/login.php", map[string]string{
				"login_username":     session.settings.User,
				"login_password":     session.settings.Password,
				"login":              "вход",
				"cap_sid":            session.captcha.Sid,
				session.captcha.Code: session.captcha.Decoded,
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
