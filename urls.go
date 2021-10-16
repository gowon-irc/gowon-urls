package main

import (
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const (
	urlRegex   = `(http(s)?:\/\/.)?(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`
	titleRegex = `<title>\n*?(.*?)\n*?</title>`
)

var (
	urlRe   = regexp.MustCompile(urlRegex)
	titleRe = regexp.MustCompile(titleRegex)
)

func extractUrls(text string) []string {
	return urlRe.FindAllString(text, -1)
}

func getBody(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

func getUrlTitle(body []byte) string {
	title := titleRe.FindString(string(body))

	if title == "" {
		return title
	}

	title = strings.Replace(title, "\n", "", -1)
	title = title[strings.Index(title, ">")+1 : strings.LastIndex(title, "<")]

	return html.UnescapeString(title)
}

func getBodys(urls []string) [][]byte {
	bodys := [][]byte{}

	for _, u := range urls {
		b, err := getBody(u)

		if err != nil {
			continue
		}

		bodys = append(bodys, b)
	}

	return bodys
}

func getTitles(bodys [][]byte) []string {
	titles := []string{}

	for _, b := range bodys {
		t := getUrlTitle(b)

		if t == "" {
			continue
		}

		titles = append(titles, t)
	}

	return titles
}

func filterUrls(urls []string, filters []*regexp.Regexp) []string {
	filterSingle := func(url string, filters []*regexp.Regexp) bool {
		for _, f := range filters {
			if f.MatchString(url) {
				return true
			}
		}

		return false
	}

	filtered := []string{}

	for _, u := range urls {
		if filterSingle(u, filters) {
			continue
		}

		filtered = append(filtered, u)
	}

	return filtered
}
