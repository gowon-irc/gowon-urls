package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Prefix  string   `short:"P" long:"prefix" env:"GOWON_PREFIX" default:"." description:"prefix for commands"`
	Broker  string   `short:"b" long:"broker" env:"GOWON_BROKER" default:"localhost:1883" description:"mqtt broker"`
	Filters []string `short:"f" long:"filters" env:"GOWON_URL_FILTERS" env-delim:"," description:"filters to apply to urls"`
}

const (
	mqttConnectRetryInternal = 5 * time.Second
)

func genUrlHandler(filters []*regexp.Regexp) func(m gowon.Message) (string, error) {
	return func(m gowon.Message) (string, error) {
		urls := extractUrls(m.Msg)
		filtered := filterUrls(urls, filters)
		bodys := getBodys(filtered)
		titles := getTitles(bodys)

		return strings.Join(titles, "\n"), nil
	}
}

func main() {
	opts := Options{}

	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	mqttOpts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", opts.Broker))
	mqttOpts.SetClientID("gowon_urls")
	mqttOpts.SetConnectRetry(true)
	mqttOpts.SetConnectRetryInterval(mqttConnectRetryInternal)

	c := mqtt.NewClient(mqttOpts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	filters := []*regexp.Regexp{}
	for _, f := range opts.Filters {
		r := regexp.MustCompile(f)
		filters = append(filters, r)
	}

	mr := gowon.NewMessageRouter()
	mr.AddRegex(urlRegex, genUrlHandler(filters))
	mr.Subscribe(c, "gowon-urls")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
}
