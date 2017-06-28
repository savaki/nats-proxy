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

