package nats_proxy

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestFromMessage(t *testing.T) {
	hKey := "X-Key"
	hValue := "value"

	cookieName := "cookie-name"
	cookieValue := "cookie-value"

	content := "hello world"
	m := &Message{
		Method: http.MethodPost,
		Header: map[string]string{
			hKey: hValue,
		},
		Cookies: map[string]*Cookie{
			cookieName: {Value: cookieValue},
		},
		Body: []byte(content),
	}
	req, err := requestFromMessage(m, "api", "api.foo.bar")
	assert.Nil(t, err)
	assert.Equal(t, "/foo/bar", req.URL.Path)

	data, err := ioutil.ReadAll(req.Body)
	assert.Nil(t, err)
	assert.Equal(t, content, string(data))
	assert.Equal(t, hValue, req.Header.Get(hKey))

	cookie, err := req.Cookie(cookieName)
	assert.Nil(t, err)
	assert.Equal(t, cookieValue, cookie.Value)
}
