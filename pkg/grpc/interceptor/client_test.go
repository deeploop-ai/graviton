package interceptor_test

import (
	"context"
	"testing"

	"github.com/deeploop-ai/orionid/internal/pkg/contexts"
	"github.com/deeploop-ai/orionid/pkg/grpc/interceptor"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestClientInfoInterceptor(t *testing.T) {
	t.Parallel()

	ic := interceptor.NewClientInfoInterceptor()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-forwarded-for", "203.0.113.1, 10.0.0.1",
		"grpcgateway-user-agent", "TestAgent/2.0",
	))

	var captured contexts.ClientInfo
	_, err := ic.UnaryMiddleware(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		captured = contexts.ClientInfoFrom(ctx)
		return nil, nil
	})
	require.NoError(t, err)
	require.Equal(t, "203.0.113.1", captured.IP)
	require.Equal(t, "TestAgent/2.0", captured.UserAgent)
}
