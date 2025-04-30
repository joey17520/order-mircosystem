package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

// TracingInterceptor 创建gRPC服务端拦截器用于追踪请求
// 返回值:
//
//	grpc.UnaryServerInterceptor - 实现了gRPC一元拦截器接口的函数
func TracingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 获取OpenTelemetry追踪器实例
		tracer := otel.Tracer("order-proxy")

		// 创建新的追踪span，使用gRPC方法全名作为span名称
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End() // 确保span在函数返回时结束

		// 调用后续处理链并返回结果
		return handler(ctx, req)
	}
}
