package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/nats-io/go-nats"
	"github.com/savaki/nats-proxy"
)

func service(label string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, label)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	nc, _ := nats.Connect(nats.DefaultURL)

	// Start Gateway
	//
	gw, _ := nats_proxy.NewGateway(
		nats_proxy.WithSubject("api"),
		nats_proxy.WithNats(nc),
	)
	server := &http.Server{
		Addr:    ":7070",
		Handler: gw,
	}
	go server.ListenAndServe()

	// Start Service Foo
	//
	foo, _ := nats_proxy.Wrap(service("foo"),
		nats_proxy.WithSubject("api.foo"), // e.g. /foo*
		nats_proxy.WithNats(nc),
	)
	doneFoo, _ := foo.Subscribe(ctx)

	// Start Service Bar
	//
	bar, _ := nats_proxy.Wrap(service("bar"),
		nats_proxy.WithSubject("api.bar"), // e.g. /bar*
		nats_proxy.WithNats(nc),
	)
	doneBar, _ := bar.Subscribe(ctx)

	// Make Request
	//
	resp, _ := http.Get("http://localhost:7070/foo")
	io.Copy(os.Stdout, resp.Body)
	io.WriteString(os.Stdout, "\n")

	resp, _ = http.Get("http://localhost:7070/bar")
	io.Copy(os.Stdout, resp.Body)
	io.WriteString(os.Stdout, "\n")

	cancel()
	<-doneFoo
	<-doneBar
}
