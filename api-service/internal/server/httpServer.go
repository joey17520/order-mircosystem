package server

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"log"
	"order-microsystem/api-service/internal/controller"
	"order-microsystem/api-service/internal/proxy"
	"order-microsystem/api-service/pkg/config"
	"order-microsystem/api-service/pkg/monitoring"
	"order-microsystem/api-service/pkg/tracing"
	"time"
)

type HTTPServer struct {
	server         *gin.Engine
	config         *config.Config
	orderProxy     *proxy.OrderProxy
	inventoryProxy *proxy.InventoryProxy
	tracer         *tracing.TracerProviderWrapper
}

func NewHTTPServer(config *config.Config, tracer *tracing.TracerProviderWrapper, logger *logrus.Logger) *HTTPServer {

	orderProxy, err := proxy.NewOrderProxy(config, logger)
	inventoryProxy, err := proxy.NewInventoryProxy(config)

	if err != nil {
		log.Fatalf("failed to create order proxy: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	server := gin.New()
	server.Use(otelgin.Middleware("api-proxy"))
	server.Use(gin.Recovery())
	server.Use(cors.New(cors.Config{
		AllowOrigins:     config.CORS.AllowOrigins,
		AllowMethods:     config.CORS.AllowMethods,
		AllowHeaders:     config.CORS.AllowHeaders,
		AllowCredentials: config.CORS.AllowCredentials,
		MaxAge:           config.CORS.MaxAge * time.Hour,
	}))

	server.GET("/metrics", gin.WrapH(promhttp.Handler()))
	server.Use(monitoring.MetricsMiddleware())

	return &HTTPServer{
		server:         server,
		config:         config,
		tracer:         tracer,
		orderProxy:     orderProxy,
		inventoryProxy: inventoryProxy,
	}
}

func (s *HTTPServer) Start() error {
	orderController := controller.NewOrderController(s.orderProxy)
	inventoryController := controller.NewInventoryController(s.inventoryProxy)

	s.server.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	api := s.server.Group("/api/v1")
	{
		api.POST("/order", orderController.CreateOrder)

		api.GET("/inventory", inventoryController.GetAllInventory)

	}
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	log.Printf("Starting HTTP Server on %s", addr)

	return s.server.Run(addr)
}
