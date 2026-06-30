package messaging

import "context"

// Mailer delivers outbound email messages.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}
