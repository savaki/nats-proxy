package nats_proxy

import "context"

// Handler publishes messages to nats and receives the response
type Handler func(ctx context.Context, subject string, message *Message) (*Message, error)

// Apply is a helper function to make calls to the Handler more readable
func (fn Handler) Apply(ctx context.Context, subject string, message *Message) (*Message, error) {
	return fn(ctx, subject, message)
}

// Nop is an empty handler that echoes the provided message; useful for testing
func Nop() Handler {
	return func(ctx context.Context, subject string, message *Message) (*Message, error) {
		return message, nil
	}
}

// Filter allows filters to be applied after nats-proxy has transformed the request into a *Message
type Filter func(Handler) Handler

// AndThen provides a convenience func to execute the handler
func (f Filter) AndThen(h Handler) Handler {
	return f(h)
}

// Chain folds a series of Filters with a Handler to construct a new Handler
func Chain(h Handler, filters ...Filter) Handler {
	for i := len(filters) - 1; i >= 0; i-- {
		h = filters[i].AndThen(h)
	}

	return func(ctx context.Context, subject string, message *Message) (*Message, error) {
		return h(ctx, subject, message)
	}
}
