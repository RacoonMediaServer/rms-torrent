package main

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"log"
	"net/url"
)

func main() {
	c := colly.NewCollector(
		colly.UserAgent("RacoonMediaServer"),
		colly.Debugger(&debug.LogDebugger{}),
	)

	err := c.Post("https://rutracker.org/forum/login.php", map[string]string{
		"login_username": "ProfessorXavier",
		"login_password": "35579007",
		"login":          "вход",
	})

	if err != nil {
		log.Fatal(err)
	}

	c.OnRequest(func(r *colly.Request) {
		log.Println("go to", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
	})

	c.OnHTML("#logged-in-username", func(e *colly.HTMLElement) {
		log.Println("found: " + e.Text)
	})

	grabber := c.Clone()

	grabber.OnRequest(func(r *colly.Request) {
		log.Println("go to", r.URL)
	})

	grabber.OnHTML("a.dl-link", func(e *colly.HTMLElement) {
		log.Println(e.Attr("href"))
	})

	grabber.OnHTML("td.message", func(e *colly.HTMLElement) {
	})

	c.OnHTML("a.tLink", func(e *colly.HTMLElement) {
		grabber.Visit("https://rutracker.org/forum/" + e.Attr("href"))
	})

	c.Visit("https://rutracker.org/forum/tracker.php?nm=" + url.QueryEscape("Соломенные еноты"))
	c.Wait()
	grabber.Wait()
}
