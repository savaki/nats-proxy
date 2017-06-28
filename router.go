package nats_proxy

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"
)

// Router provides a wrapper over the standard http.Handler interface and acts as a bridge between the http.Handler and
// nats
type Router struct {
	h              http.Handler
	nc             *nats.Conn
	subject        string
	returnNotFound bool
}

// Wrap an existing http.Handler with the specified options
func Wrap(h http.Handler, opts ...Option) (*Router, error) {
	c, err := readConfig(opts...)
	if err != nil {
		return nil, err
	}

	r := &Router{
		h:              h,
		nc:             c.nc,
		subject:        c.subject,
		returnNotFound: c.returnNotFound,
	}

	return r, nil
}

// Subscribe listens to the subject specified
func (r *Router) Subscribe(ctx context.Context) (<-chan struct{}, error) {
	subject := r.subject
	for strings.HasSuffix(subject, ".") {
		subject = subject[0 : len(subject)-1]
	}

	root, err := r.nc.QueueSubscribe(subject, "worker", r.handler)
	if err != nil {
		return nil, err
	}

	children, err := r.nc.QueueSubscribe(subject+".>", "worker", r.handler)
	if err != nil {
		root.Unsubscribe()
		return nil, err
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		defer root.Unsubscribe()
		defer children.Unsubscribe()
		<-ctx.Done()
	}()

	return done, nil
}

func (r *Router) handler(msg *nats.Msg) {
	req, err := requestFromMessage(msg, r.subject)
	if err != nil {
		log.Fatalln(err)
	}

	w := httptest.NewRecorder()
	r.h.ServeHTTP(w, req)

	if r.returnNotFound && msg.Reply != "" {
		writeResponse(r.nc, msg.Reply, w)
	}
}

func requestFromMessage(msg *nats.Msg, rootSubject string) (*http.Request, error) {
	m := &Message{}
	if err := proto.Unmarshal(msg.Data, m); err != nil {
		return nil, err
	}

	var body io.Reader
	if msg.Data != nil {
		body = bytes.NewReader(msg.Data)
	}

	subject := msg.Subject
	if strings.HasPrefix(subject, rootSubject) {
		subject = subject[len(rootSubject):]
	}
	for strings.HasPrefix(subject, ".") {
		subject = subject[1:]
	}
	subject = strings.Replace(subject, ".", "/", -1)

	req, err := http.NewRequest(m.Method, "http://localhost/"+subject, body)
	if err != nil {
		return nil, err
	}

	for k, v := range m.Header {
		req.Header.Set(k, v)
	}

	for name, cookie := range m.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: cookie.Value,
			Path:  cookie.Path,
		})
	}

	return req, nil
}

func writeResponse(nc *nats.Conn, subject string, w *httptest.ResponseRecorder) {
	m := &Message{
		Status: int32(w.Code),
		Header: map[string]string{},
	}

	for key := range w.HeaderMap {
		m.Header[key] = w.HeaderMap.Get(key)
	}

	if w.Body != nil {
		m.Body = w.Body.Bytes()
	}

	data, err := proto.Marshal(m)
	if err != nil {
		log.Printf("Unable to marshal message, %v\n", err)
	}

	err = nc.Publish(subject, data)
	if err != nil {
		log.Printf("Unable to publish message, %v\n", err)
	}
}
