package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/savaki/nats-proxy"
	"github.com/urfave/cli"
)

type options struct {
	Port    int
	Subject string
	Headers string
	Cookies string
}

var opts options

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "port",
			Value:       5050,
			Usage:       "port to listen on",
			EnvVar:      "PORT",
			Destination: &opts.Port,
		},
		cli.StringFlag{
			Name:        "subject",
			Usage:       "subject root that all messages will be prefixed with",
			EnvVar:      "SUBJECT",
			Destination: &opts.Subject,
		},
		cli.StringFlag{
			Name:        "headers",
			Usage:       "comma separated list of HTTP headers to pass through",
			EnvVar:      "HEADERS",
			Destination: &opts.Headers,
		},
		cli.StringFlag{
			Name:        "cookies",
			Usage:       "comma separated list of HTTP cookies to pass through",
			EnvVar:      "COOKIES",
			Destination: &opts.Cookies,
		},
	}
	app.Action = run
	app.Run(os.Args)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func run(_ *cli.Context) error {
	proxy, err := nats_proxy.NewGateway(
		nats_proxy.WithSubject(opts.Subject),
		nats_proxy.WithHeaders(strings.Split(opts.Headers, ",")...),
		nats_proxy.WithCookies(strings.Split(opts.Cookies, ",")...),
	)
	check(err)

	fmt.Printf("Listening on port %v\n", opts.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", opts.Port), proxy)
	check(err)

	return nil
}
