package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/altairsix/nats-proxy"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
)

type Options struct {
	Subject string
}

var opts Options

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "subject",
			Value:       "api.sample",
			Usage:       "mount point for service",
			EnvVar:      "SUBJECT",
			Destination: &opts.Subject,
		},
	}
	app.Action = Run
	app.Run(os.Args)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func Run(_ *cli.Context) error {
	h := gin.New()
	h.POST("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})

	proxy, err := nats_proxy.Wrap(h,
		nats_proxy.WithSubject(opts.Subject),
	)
	check(err)

	done, err := proxy.Subscribe(context.Background())
	check(err)

	<-done

	return nil
}
