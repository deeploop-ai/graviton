package contexts

import (
	"context"

	"github.com/deeploop-ai/fleet/internal/domain/shared"
)

func WithPrincipal(ctx context.Context, p *shared.Principal) context.Context {
	return context.WithValue(ctx, ContextKeyPrincipal, p)
}

func Principal(ctx context.Context) (*shared.Principal, bool) {
	v := ctx.Value(ContextKeyPrincipal)
	p, ok := v.(*shared.Principal)
	return p, ok && p != nil
}

func WithProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, ContextKeyProjectID, projectID)
}

func ProjectID(ctx context.Context) (string, bool) {
	v := ctx.Value(ContextKeyProjectID)
	s, ok := v.(string)
	return s, ok && s != ""
}
