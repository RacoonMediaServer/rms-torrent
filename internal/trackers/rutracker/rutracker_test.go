package rutracker

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"git.rms.local/RacoonMediaServer/rms-torrent/internal/types"
)

const captchaWaitTime = 40 * time.Second

func waitCaptcha(url string, t *testing.T) string {
	fmt.Printf("Please enter captcha text for %s\n", url)
	<-time.After(captchaWaitTime)
	data, err := ioutil.ReadFile("captcha.txt")
	if err != nil {
		t.Error(err)
	}

	return string(data)
}

func TestInvalidAuth(t *testing.T) {
	session := SearchSession{}
	session.Setup(types.SessionSettings{
		User:      "xxxxxzz",
		Password:  "yuuuuzxyy",
		UserAgent: "RacoonMediaServer",
	})

	torrents, err := session.Search("Матрица")

	if err == nil {
		t.Error("Error must be raised")
	}

	e, success := err.(types.Error)
	if !success {
		t.Error("Error must be types.Error")
	}

	if torrents != nil {
		t.Error("Torrents must be nil")
	}

	if !(e.Code == types.AuthFailed || e.Code == types.CaptchaRequired) {
		t.Errorf("Error must be AuthFailed, e = [%d] %s", e.Code, e.Error())
	}
}

func TestSuccessAuth(t *testing.T) {
	session := SearchSession{}
	session.Setup(types.SessionSettings{
		User:      "ProfessorXavier",
		Password:  "35579007",
		UserAgent: "RacoonMediaServer",
	})

	torrents, err := session.Search("Матрица")

	e, _ := err.(types.Error)
	if e.Code == types.CaptchaRequired {
		session.SetCaptchaText(waitCaptcha(e.Captcha, t))
		torrents, err = session.Search("Матрица")
	}

	if len(torrents) == 0 {
		t.Error("Result must be not nil")
	}

	if err != nil {
		t.Errorf("Error must be nil: %+s", err.Error())
	}

	_, err = session.Download(torrents[0].DownloadLink)
	if err != nil {
		t.Errorf("Download failed: %+s", err.Error())
	}
}
