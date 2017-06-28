package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/savaki/nats-proxy"
	"github.com/nats-io/go-nats"
)

func Hello(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "hello world")
}

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
	r, _ := nats_proxy.Wrap(http.HandlerFunc(Hello),
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
