package monitoring

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		// 修改为4个标签值：服务名、方法名、状态码、状态
		statusCode := "OK"  // 默认值
		status := "success" // 默认值

		resp, err = handler(ctx, req)

		if err != nil {
			statusCode = "Unknown"
			status = "error"
		}

		RequestCount.WithLabelValues(
			"inventory-service", // 服务名
			info.FullMethod,     // 方法名
			statusCode,          // 状态码
			status,              // 状态
		).Inc()

		RequestDuration.WithLabelValues(
			"inventory-service",
			info.FullMethod,
			statusCode,
		).Observe(time.Since(start).Seconds())

		return resp, err
	}
}
