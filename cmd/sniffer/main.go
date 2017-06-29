package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/savaki/nats-proxy"
	"github.com/urfave/cli"
)

type Options struct {
	Subject string
	NatsURL string
	Queue   string
}

var opts Options

func main() {
	buf := make([]byte, 12)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Read(buf)

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
			Destination: &opts.NatsURL,
		},
		cli.StringFlag{
			Name:        "queue",
			Value:       hex.EncodeToString(buf),
			Usage:       "unique queue name",
			Destination: &opts.Queue,
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
	nc, err := nats.Connect(opts.NatsURL)
	check(err)

	h := Dump()
	r, err := nats_proxy.Wrap(h,
		nats_proxy.WithSubject(opts.Subject),
		nats_proxy.WithNats(nc),
		nats_proxy.WithQueue(opts.Queue),
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
