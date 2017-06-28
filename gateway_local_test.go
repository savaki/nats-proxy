package nats_proxy

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	filter := func(status int32) Filter {
		return func(h Handler) Handler {
			return func(ctx context.Context, subject string, message *Message) (*Message, error) {
				out, _ := h(ctx, subject, message)
				out.Status = status
				return out, nil
			}
		}
	}

	status := int32(http.StatusUnauthorized)
	gw, err := NewGateway(
		WithFilters(filter(status)),
		WithNopHandler(),
	)
	assert.Nil(t, err)

	out, err := gw.h.Apply(context.Background(), "blah", &Message{})
	assert.Nil(t, err)
	assert.Equal(t, status, out.Status)
}
