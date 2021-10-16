package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func genHTTPServer(body string) *httptest.Server {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	})
	return httptest.NewServer(h)
}

var (
	server1 = genHTTPServer("no title")
	server2 = genHTTPServer("")
)

func TestUrlRegex(t *testing.T) {
	cases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "no schema no www",
			url:      "website.com",
			expected: true,
		},
		{
			name:     "no schema www",
			url:      "www.website.com",
			expected: true,
		},
		{
			name:     "http no www",
			url:      "http://website.com",
			expected: true,
		},
		{
			name:     "http www",
			url:      "http://www.website.com",
			expected: true,
		},
		{
			name:     "https no www",
			url:      "https://website.com",
			expected: true,
		},
		{
			name:     "https www",
			url:      "https://www.website.com",
			expected: true,
		},
		{
			name:     "no url",
			url:      "hello i am a string",
			expected: false,
		},
		{
			name:     "url in string",
			url:      "test website.com test",
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			match, _ := regexp.MatchString(urlRegex, tc.url)
			assert.Equal(t, match, tc.expected)
		})
	}
}

func TestExtractUrls(t *testing.T) {
	cases := []struct {
		name string
		text string
		urls []string
	}{
		{
			name: "only url",
			text: "website.com",
			urls: []string{"website.com"},
		},
		{
			name: "multiple urls",
			text: "website.com test.com",
			urls: []string{"website.com", "test.com"},
		},
		{
			name: "one url in string",
			text: "go to website.com",
			urls: []string{"website.com"},
		},
		{
			name: "multiple urls in string",
			text: "go to website.com and test.com",
			urls: []string{"website.com", "test.com"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urls := extractUrls(tc.text)

			assert.Equal(t, tc.urls, urls)
		})
	}
}

func TestGetBody(t *testing.T) {
	cases := []struct {
		name       string
		url        string
		body       string
		returnsErr bool
	}{
		{
			name:       "valid url",
			url:        server1.URL,
			body:       "no title",
			returnsErr: false,
		},
		{
			name:       "not a url",
			url:        "test",
			body:       "",
			returnsErr: true,
		},
		{
			name:       "failing url",
			url:        "127.0.0.1:0",
			body:       "",
			returnsErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := getBody(tc.url)

			if tc.returnsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tc.body, string(body))
		})
	}
}

func TestUrlTitle(t *testing.T) {
	cases := []struct {
		name  string
		body  []byte
		title string
	}{
		{
			name:  "no title",
			body:  []byte("test body"),
			title: "",
		},
		{
			name:  "has title",
			body:  []byte("<title>title</title>"),
			title: "title",
		},
		{
			name:  "title with newline",
			body:  []byte("<title>\ntitle\n</title>"),
			title: "title",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			title := getUrlTitle(tc.body)

			assert.Equal(t, tc.title, title)
		})
	}
}

func TestGetBodys(t *testing.T) {
	cases := []struct {
		name  string
		urls  []string
		bodys [][]byte
	}{
		{
			name:  "one url",
			urls:  []string{server1.URL},
			bodys: [][]byte{[]byte("no title")},
		},
		{
			name:  "two urls",
			urls:  []string{server1.URL, server2.URL},
			bodys: [][]byte{[]byte("no title"), []byte("")},
		},
		{
			name:  "no urls",
			urls:  []string{},
			bodys: [][]byte{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bodys := getBodys(tc.urls)
			assert.Equal(t, tc.bodys, bodys)
		})
	}
}

func TestGetTitles(t *testing.T) {
	cases := []struct {
		name   string
		bodys  [][]byte
		titles []string
	}{
		{
			name:   "one title",
			bodys:  [][]byte{[]byte("<title>title</title>")},
			titles: []string{"title"},
		},
		{
			name:   "two titles",
			bodys:  [][]byte{[]byte("<title>title</title>"), []byte("<title>title2</title>")},
			titles: []string{"title", "title2"},
		},
		{
			name:   "one title one no title",
			bodys:  [][]byte{[]byte("<title>title</title>"), []byte("content")},
			titles: []string{"title"},
		},
		{
			name:   "none",
			bodys:  [][]byte{},
			titles: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			titles := getTitles(tc.bodys)

			assert.Equal(t, tc.titles, titles)
		})
	}
}

func TestFilterUrls(t *testing.T) {
	filter1 := regexp.MustCompile(`.*website.com.*`)
	filter2 := regexp.MustCompile(`.*test.com.*`)

	url1 := "www.website.com/1"
	url2 := "http://test.com/2"
	url3 := "https://www.url.com/3"

	cases := []struct {
		name     string
		urls     []string
		filters  []*regexp.Regexp
		filtered []string
	}{
		{
			name:     "one url one filter no results",
			urls:     []string{url1},
			filters:  []*regexp.Regexp{filter1},
			filtered: []string{},
		},
		{
			name:     "one url one filter one result",
			urls:     []string{url1},
			filters:  []*regexp.Regexp{filter2},
			filtered: []string{url1},
		},
		{
			name:     "two urls one filter one result",
			urls:     []string{url1, url2},
			filters:  []*regexp.Regexp{filter1},
			filtered: []string{url2},
		},
		{
			name:     "one url no filter one result",
			urls:     []string{url1},
			filters:  []*regexp.Regexp{},
			filtered: []string{url1},
		},
		{
			name:     "three urls two filters one result",
			urls:     []string{url1, url2, url3},
			filters:  []*regexp.Regexp{filter1, filter2},
			filtered: []string{url3},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filtered := filterUrls(tc.urls, tc.filters)
			assert.Equal(t, tc.filtered, filtered)
		})
	}
}
