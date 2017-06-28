package nats_proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/altairsix/nats-proxy"
	"github.com/stretchr/testify/assert"
)

func TestNewProxy(t *testing.T) {
	p, err := nats_proxy.NewGateway(
		nats_proxy.WithNopHandler(),
	)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost", nil)

	// When
	p.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)
}
