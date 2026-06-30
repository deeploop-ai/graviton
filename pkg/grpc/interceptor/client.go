package interceptor

import (
	"context"
	"strings"

	"github.com/deeploop-ai/orionid/internal/pkg/contexts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientInfoInterceptor extracts client IP and user agent from gRPC metadata
// (populated by grpc-gateway from HTTP headers) into the request context.
type ClientInfoInterceptor struct{}

func NewClientInfoInterceptor() *ClientInfoInterceptor {
	return &ClientInfoInterceptor{}
}

func (c *ClientInfoInterceptor) UnaryMiddleware(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ctx = contexts.WithClientInfo(ctx, extractClientInfo(md))
	}
	return handler(ctx, req)
}

func extractClientInfo(md metadata.MD) contexts.ClientInfo {
	ip := firstMetadataValue(md, "x-forwarded-for")
	if ip != "" {
		if idx := strings.Index(ip, ","); idx > 0 {
			ip = strings.TrimSpace(ip[:idx])
		}
	}
	if ip == "" {
		ip = firstMetadataValue(md, "x-real-ip")
	}
	ua := firstMetadataValue(md, "grpcgateway-user-agent")
	if ua == "" {
		ua = firstMetadataValue(md, "user-agent")
	}
	return contexts.ClientInfo{IP: ip, UserAgent: ua}
}
