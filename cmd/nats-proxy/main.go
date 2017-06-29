package main

import (
	"context"
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
	Set     cli.StringSlice
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
		cli.StringSliceFlag{
			Name:  "set",
			Usage: "set header items KEY=VALUE",
			Value: &opts.Set,
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

func SetHeaders() nats_proxy.Filter {
	return func(h nats_proxy.Handler) nats_proxy.Handler {
		return func(ctx context.Context, subject string, message *nats_proxy.Message) (*nats_proxy.Message, error) {
			for _, item := range opts.Set {
				if segments := strings.SplitN(item, "=", 2); len(segments) == 2 {
					key := segments[0]
					value := segments[1]
					message.Header[key] = value
				}
			}
			return h(ctx, subject, message)
		}
	}
}

func run(_ *cli.Context) error {
	proxy, err := nats_proxy.NewGateway(
		nats_proxy.WithSubject(opts.Subject),
		nats_proxy.WithHeaders(strings.Split(opts.Headers, ",")...),
		nats_proxy.WithCookies(strings.Split(opts.Cookies, ",")...),
		nats_proxy.WithFilters(SetHeaders()),
	)
	check(err)

	fmt.Printf("Listening on port %v\n", opts.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", opts.Port), proxy)
	check(err)

	return nil
}
