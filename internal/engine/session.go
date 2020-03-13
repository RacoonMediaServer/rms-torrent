package engine

import "github.com/gocolly/colly/v2"

type Session struct {
	c *colly.Collector
}

func (session *Session) Login(user,password string) error {
	return nil
}

func (session *Session) AcceptCaptcha(text string) error {
	return nil
}


