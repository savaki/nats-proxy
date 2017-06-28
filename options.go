package nats_proxy

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
)

const (
	// DefaultSubject provides the name of the root subject nats-proxy will bind to
	DefaultSubject = "api"

	// DefaultTimeout specifies the maximum amount of time a request may take
	DefaultTimeout = time.Second * 10

	// DefaultQueue specifies the name of the queue for Conn.QueueSubscribe
	DefaultQueue = "service"
)

type config struct {
	handler        Handler
	url            string
	queue          string
	nc             *nats.Conn
	filters        []Filter
	headers        map[string]struct{}
	cookies        map[string]struct{}
	subject        string
	timeout        time.Duration
	returnNotFound bool
	onError        func(err error, w http.ResponseWriter, req *http.Request)
}

type Option func(*config)

// WithNats specifies the nats connection to use
func WithNats(nc *nats.Conn) Option {
	return func(p *config) {
		p.nc = nc
	}
}

// WithNats specifies the nats url to use
func WithUrl(url string) Option {
	return func(p *config) {
		p.url = url
	}
}

// WithQueue specifies the queue the nats subscribers should belong to
func WithQueue(queue string) Option {
	return func(p *config) {
		p.queue = queue
	}
}

// WithSubject specifies the nats subject root to publish to
func WithSubject(subject string) Option {
	return func(p *config) {
		if subject != "" {
			p.subject = subject
		}
	}
}

// WithHeaders specifies http headers that should be passed across nats
func WithHeaders(headers ...string) Option {
	return func(p *config) {
		for _, item := range headers {
			if v := strings.TrimSpace(item); v != "" {
				p.headers[v] = struct{}{}
			}
		}
	}
}

// WithFilters allows gateway filters to be specified; applies ONLY to Gateway
func WithFilters(filters ...Filter) Option {
	return func(p *config) {
		p.filters = append(p.filters, filters...)
	}
}

// WithCookies specifies the specific cookies that should be passed across nats
func WithCookies(cookies ...string) Option {
	return func(p *config) {
		for _, item := range cookies {
			if v := strings.TrimSpace(item); v != "" {
				p.cookies[v] = struct{}{}
			}
		}
	}
}

// WithTimeout specifies how long a single request can take; defaults to ```nats_proxy.DefaultTimeout```
func WithTimeout(d time.Duration) Option {
	return func(p *config) {
		p.timeout = d
	}
}

// WithNotFoundEnabled specifies whether or not the ```*nats_proxy.Router``` should respond with 404 Not Found for
// routes it's unable to handle
func WithNotFoundEnabled(enabled bool) Option {
	return func(p *config) {
		p.returnNotFound = enabled
	}
}

// WithNopHandler provides an nop handler useful for testing; the content submitted will be echoed back
func WithNopHandler() Option {
	return func(p *config) {
		p.handler = func(ctx context.Context, subject string, message *Message) (*Message, error) {
			return message, nil
		}
	}
}

// WithHandler allows the underlying handler to be specified; useful primarily for testing
func WithHandler(h Handler) Option {
	return func(p *config) {
		p.handler = h
	}
}

func readConfig(opts ...Option) (*config, error) {
	c := &config{
		subject:        DefaultSubject,
		queue:          DefaultQueue,
		timeout:        DefaultTimeout,
		onError:        onError,
		returnNotFound: true,
		headers: map[string]struct{}{
			"Content-Disposition": {},
			"Content-Encoding":    {},
			"Content-Length":      {},
			"Content-Range":       {},
			"Content-Type":        {},
			"Set-Cookie":          {},
		},
		cookies: map[string]struct{}{},
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.nc == nil {
		if c.url == "" {
			c.url = nats.DefaultURL
		}

		v, err := nats.Connect(c.url)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to connect to nats url, %v", c.url)
		}
		c.nc = v
	}

	return c, nil
}
