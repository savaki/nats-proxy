package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"

	"github.com/nats-io/go-nats"
	"github.com/savaki/nats-proxy"
	"github.com/urfave/cli"
)

type Options struct {
	Subject string
	NatsUrl string
}

var opts Options

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "subject",
			Value:       "api",
			Usage:       "root subject to listen to",
			Destination: &opts.Subject,
		},
		cli.StringFlag{
			Name:        "server",
			Value:       nats.DefaultURL,
			Usage:       "nats server to connect to",
			Destination: &opts.NatsUrl,
		},
	}
	app.Action = Run
	app.Run(os.Args)
}

func Dump() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		data, err := httputil.DumpRequest(req, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println()
		fmt.Println(string(data))

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	}
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func Run(_ *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	nc, err := nats.Connect(opts.NatsUrl)
	check(err)

	h := Dump()
	r, err := nats_proxy.Wrap(h,
		nats_proxy.WithSubject(opts.Subject),
		nats_proxy.WithNats(nc),
	)
	check(err)

	done, err := r.Subscribe(ctx)
	check(err)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	cancel()
	<-done

	return nil
}
