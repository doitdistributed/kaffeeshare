package email

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"

	"github.com/mvdan/xurls"
)

func parseBody(c context.Context, mail *email) ([]string, error) {
	if mail.ContentType[:4] == "html" {
		return parseHTMLBody(c, mail.Body)
	}

	if mail.ContentType[:4] == "text" {
		return parseTextBody(c, mail.Body)
	}

	return nil, fmt.Errorf("Unsupported content type: %s", mail.ContentType)
}

func parseHTMLBody(c context.Context, body string) ([]string, error) {
	return firstURLFromHTML(c, body)
}

func firstURLFromHTML(c context.Context, body string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// use a 'set' to remove duplicates
	set := make(map[string]bool)
	doc.Find("a").First().Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}
		set[link] = true

		log.Infof(c, "HTML found %v", link)
	})

	links := make([]string, len(set))
	i := 0
	for k := range set {
		links[i] = k
		i++
	}

	return links, nil
}

func allURLsFromHTML(c context.Context, body string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// use a 'set' to remove duplicates
	set := make(map[string]bool)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}
		set[link] = true

		log.Infof(c, "HTML found %v", link)
	})

	links := make([]string, len(set))
	i := 0
	for k := range set {
		links[i] = k
		i++
	}

	return links, nil
}

func parseTextBody(c context.Context, body string) ([]string, error) {
	return firstURLFromText(c, body)
}

func firstURLFromText(c context.Context, body string) ([]string, error) {
	links := make([]string, 1)
	links[0] = xurls.Relaxed.FindString(body)
	log.Infof(c, "Found urls in body %v,  %v", body, links)

	return links, nil
}

func allURLsFromText(c context.Context, body string) ([]string, error) {
	links := xurls.Relaxed.FindAllString(body, -1)
	log.Infof(c, "Found urls in body %v,  %v", body, links)

	set := make(map[string]bool)
	for _, l := range links {
		set[l] = true
	}

	links = make([]string, len(set))
	i := 0
	for k := range set {
		links[i] = k
		i++
	}

	return links, nil
}
