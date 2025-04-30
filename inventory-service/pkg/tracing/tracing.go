package tracing

import (
	"context"
	"fmt"
	"order-microsystem/inventory-service/pkg/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type TracerProviderWrapper struct {
	*sdktrace.TracerProvider
}

// InitTracer 初始化OpenTelemetry追踪器
// 参数:
//
//	cfg - 应用配置，包含Jaeger相关配置信息
//
// 返回值:
//
//	*TracerProviderWrapper - 包装后的追踪器提供者
//	error - 如果初始化过程中出现错误则返回错误信息
func InitTracer(cfg *config.Config) (*TracerProviderWrapper, error) {
	// 构建Jaeger收集器端点URL
	endpoint := fmt.Sprintf("http://%s:%d/api/traces", cfg.Jaeger.AgentHost, cfg.Jaeger.AgentPort)

	// 创建Jaeger导出器
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// 创建资源描述信息
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.Jaeger.ServiceName), // 设置服务名称
			semconv.DeploymentEnvironment("production"), // 设置环境为生产环境
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建追踪器提供者
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),                // 使用批处理导出器
		sdktrace.WithResource(res),                    // 设置资源信息
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // 设置采样策略为全采样
	)

	// 设置全局追踪器提供者
	otel.SetTracerProvider(tp)

	// 设置全局文本映射传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // 支持W3C TraceContext标准
		propagation.Baggage{},      // 支持W3C Baggage标准
	))

	// 返回包装后的追踪器提供者
	return &TracerProviderWrapper{tp}, nil
}
