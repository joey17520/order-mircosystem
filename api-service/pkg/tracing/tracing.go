package tracing

import (
	"context"
	"fmt"
	"log"
	"order-microsystem/api-service/pkg/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type TracerProviderWrapper struct {
	*sdktrace.TracerProvider
}

func InitTracer(cfg *config.Config) (*TracerProviderWrapper, error) {
	endpoint := fmt.Sprintf("http://%s:%d/api/traces", cfg.Jaeger.AgentHost, cfg.Jaeger.AgentPort)
	log.Printf("Connecting to Jaeger at: %s", endpoint) // 添加连接日志

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		log.Printf("Failed to create Jaeger exporter: %v", err) // 更详细的错误日志
		return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.Jaeger.ServiceName),
			semconv.DeploymentEnvironment("production"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracerProviderWrapper{tp}, nil

}
