package nats_proxy

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"
)

// Gateway is our http -> nats gateway
type Gateway struct {
	headers map[string]struct{}
	cookies map[string]struct{}
	subject string
	h       Handler
	onError func(err error, w http.ResponseWriter, req *http.Request)
}

// ServeHTTP implements the http.Handler contract.  Wraps messages into a *Message and performs a nats request
func (p *Gateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	subject := makeSubject(req, p.subject)

	in, err := messageFromRequest(req, p.headers, p.cookies)
	if err != nil {
		p.onError(err, w, req)
		return
	}

	out, err := p.h.Apply(req.Context(), subject, in)
	if err != nil {
		p.onError(err, w, req)
		return
	}

	writeMessage(w, out)
}

// NewGateway returns a new http to nats gateway with the options provided
func NewGateway(opts ...Option) (*Gateway, error) {
	c, err := readConfig(opts...)
	if err != nil {
		return nil, err
	}

	h := c.handler
	if h == nil {
		h = request(c.nc, c.timeout)
	}

	return &Gateway{
		headers: c.headers,
		subject: c.subject,
		h:       h,
		onError: c.onError,
	}, nil
}

func makeSubject(req *http.Request, prefix string) string {
	path := req.URL.Path
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	path = strings.Replace(path, "/", ".", -1)

	subject := prefix + "." + path
	for strings.HasSuffix(subject, ".") {
		subject = subject[0 : len(subject)-1]
	}
	return subject
}

func messageFromRequest(req *http.Request, headers, cookies map[string]struct{}) (*Message, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	h := map[string]string{}
	for key := range headers {
		if value := req.Header.Get(key); value != "" {
			h[key] = value
		}
	}

	c := map[string]*Cookie{}
	for name := range cookies {
		if cookie, err := req.Cookie(name); err == nil {
			c[name] = &Cookie{
				Value: cookie.Value,
				Path:  cookie.Path,
			}
		}
	}

	return &Message{
		Method:  req.Method,
		Header:  h,
		Cookies: c,
		Body:    body,
	}, nil
}

func writeMessage(w http.ResponseWriter, out *Message) {
	for k, v := range out.Header {
		w.Header().Set(k, v)
	}
	status := out.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(int(status))

	if out.Body != nil {
		w.Write(out.Body)
	}
}

func request(nc *nats.Conn, timeout time.Duration) Handler {
	return func(ctx context.Context, subject string, m *Message) (*Message, error) {
		data, err := proto.Marshal(m)
		if err != nil {
			return nil, err
		}

		if timeout > 0 {
			ctx, _ = context.WithTimeout(ctx, timeout)
		}
		out, err := nc.RequestWithContext(ctx, subject, data)
		if err != nil {
			return nil, err
		}

		outMessage := &Message{}
		if err := proto.Unmarshal(out.Data, outMessage); err != nil {
			return nil, err
		}

		return outMessage, nil
	}
}

func onError(err error, w http.ResponseWriter, _ *http.Request) {
	content := ""
	if err != nil {
		content = err.Error()
	}
	http.Error(w, content, http.StatusBadRequest)
}
