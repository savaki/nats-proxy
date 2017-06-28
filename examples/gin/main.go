package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/go-nats"
	"github.com/savaki/nats-proxy"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	nc, _ := nats.Connect(nats.DefaultURL)

	// Start Gateway
	//
	gw, _ := nats_proxy.NewGateway(nats_proxy.WithNats(nc))
	server := &http.Server{
		Addr:    ":7070",
		Handler: gw,
	}
	go server.ListenAndServe()

	// Start Service
	//
	h := gin.New()
	h.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})
	r, _ := nats_proxy.Wrap(h,
		nats_proxy.WithNats(nc),
	)
	done, _ := r.Subscribe(ctx)

	// Make Request
	//
	resp, _ := http.Get("http://localhost:7070/")
	io.Copy(os.Stdout, resp.Body)

	cancel()
	<-done
}
