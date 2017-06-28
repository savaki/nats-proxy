package nats_proxy_test

import (
	"context"
	"testing"

	"github.com/savaki/nats-proxy"
	"github.com/stretchr/testify/assert"
)

func TestChain(t *testing.T) {
	callOrder := []string{}

	service := func(label string) nats_proxy.Filter {
		return func(h nats_proxy.Handler) nats_proxy.Handler {
			return func(ctx context.Context, subject string, message *nats_proxy.Message) (*nats_proxy.Message, error) {
				callOrder = append(callOrder, label)
				return h(ctx, subject, message)
			}
		}
	}

	h := nats_proxy.Chain(nats_proxy.Nop(), service("a"), service("b"), service("c"))
	_, err := h.Apply(context.Background(), "blah", nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, callOrder)
}
