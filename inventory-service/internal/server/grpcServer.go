package server

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"order-microsystem/inventory-service/internal/controller"
	"order-microsystem/inventory-service/pkg/config"
	"order-microsystem/inventory-service/pkg/monitoring"
	"order-microsystem/inventory-service/pkg/tracing"
	"time"
)

type GRPCServer struct {
	server *grpc.Server
	config *config.Config
	consul *api.Client
}

func NewGRPCServer(config *config.Config) *GRPCServer {
	opts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
		)),
		grpc.ChainUnaryInterceptor(
			tracing.TracingInterceptor(),
			monitoring.UnaryServerInterceptor(),
		),
	}

	return &GRPCServer{
		server: grpc.NewServer(opts...),
		config: config,
	}
}

func (s *GRPCServer) Start(inventorycontroller *controller.InventoryController) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	controller.RegisterInventoryController(s.server, inventorycontroller)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s.server, healthServer)

	reflection.Register(s.server)

	// register with consul
	if err := s.RegisterWithConsul(); err != nil {
		return fmt.Errorf("failed to register with consul: %v", err)
	}

	log.Printf("starting gRPC Server on %s:%d", s.config.Server.Host, s.config.Server.Port)

	go func() {
		if err := s.server.Serve(listen); err != nil {
			fmt.Printf("failed to serve: %v", err)
		}
	}()

	//添加HTTP metrics服务器
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsAddr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.MetricsPort)
		log.Printf("Starting metrics server on %s", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			log.Printf("Metrics server failed: %v", err)
		}
	}()
	return nil
}

func (s *GRPCServer) RegisterWithConsul() error {
	consulClient := api.DefaultConfig()
	consulClient.Address = fmt.Sprintf("%s:%d", s.config.Consul.Host, s.config.Consul.Port)
	// 添加重试逻辑
	var client *api.Client
	var err error
	for i := 0; i < 3; i++ {
		client, err = api.NewClient(consulClient)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	s.consul = client

	registration := &api.AgentServiceRegistration{
		ID:      s.config.Consul.ServiceID,
		Name:    s.config.Consul.ServiceName,
		Address: s.config.Consul.ServiceName,
		Port:    s.config.Server.Port,
		Check: &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d", s.config.Consul.ServiceName, s.config.Server.Port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := s.consul.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register proxy with Consul: %v", err)
	}

	log.Println("Successfully registered with Consul")
	return nil
}

func (s *GRPCServer) Shutdown() {
	log.Printf("shutting down gRPC server gracefully...")

	if s.consul != nil {
		if err := s.consul.Agent().ServiceDeregister(s.config.Consul.ServiceID); err != nil {
			log.Printf("failed to deregister proxy with Consul: %v", err)
		} else {
			log.Println("Successfully deregistered with Consul")
		}
	}

	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	timer := time.NewTimer(10 * time.Second)
	select {
	case <-timer.C:
		// 超时后强制停止
		s.server.Stop()
		log.Fatal("gRPC server stopped forcefully due to timeout")
	case <-stopped:
		timer.Stop()
		log.Fatal("gRPC server stopped gracefully")
	}
}
