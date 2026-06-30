package messaging

import "context"

// SMSSender delivers outbound SMS messages.
type SMSSender interface {
	Send(ctx context.Context, to, body string) error
}
