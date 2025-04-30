package main

import (
	"context"
	"fmt"
	"log"
	"order-microsystem/order-service/internal/controller"
	"order-microsystem/order-service/internal/domain/repository/mongodb"
	"order-microsystem/order-service/internal/server"
	"order-microsystem/order-service/internal/service"
	"order-microsystem/order-service/pkg/cache"
	"order-microsystem/order-service/pkg/config"
	"order-microsystem/order-service/pkg/database"
	"order-microsystem/order-service/pkg/messaging"
	"order-microsystem/order-service/pkg/tracing"
	"os"
	"os/signal"
	"syscall"
)

// main 是程序的入口函数，负责初始化系统的各个组件，
// 包括配置、追踪系统、数据库、消息队列、缓存等，
// 启动 gRPC 服务器，并处理优雅关闭逻辑。
func main() {
	// 加载配置文件，配置文件名为 "config"
	cfg, err := config.NewConfig("config")
	if err != nil {
		// 若配置文件加载失败，记录错误信息并终止程序
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化分布式追踪系统，用于分布式系统的性能监控和问题排查
	tracerProvider, err := tracing.InitTracer(cfg)
	if err != nil {
		// 若追踪系统初始化失败，记录错误信息并终止程序
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	// 延迟关闭追踪系统，确保程序结束时能正确释放资源
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			// 若关闭追踪系统失败，记录错误信息并终止程序
			log.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	}()

	// 构建 MongoDB 连接 URI
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", cfg.Database.Mongo.Username, cfg.Database.Mongo.Password, cfg.Database.Mongo.Host, cfg.Database.Mongo.Port)
	// 连接 MongoDB 数据库
	mongoClient, err := database.NewMongoDB(uri)
	if err != nil {
		// 若连接数据库失败，记录错误信息并终止程序
		log.Fatal(err.Error())
	}
	// 延迟断开 MongoDB 连接，确保程序结束时能正确关闭数据库连接
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			// 若断开连接失败，记录错误信息并终止程序
			log.Fatalf("failed to close MongoDB: %v", err)
		}
	}()
	// 获取指定名称的 MongoDB 数据库实例
	db := mongoClient.Client.Database(cfg.Database.Mongo.Database)
	// 创建订单仓库实例，用于操作 MongoDB 中的订单数据
	orderRepo := mongodb.NewOrderRepository(db)

	// 连接 RabbitMQ 消息队列，传入配置信息和订单仓库实例
	rabbitmq, err := messaging.NewRabbitMQ(&cfg.RabbitMQ, orderRepo)
	if err != nil {
		// 若连接 RabbitMQ 失败，记录错误信息并终止程序
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	// 延迟关闭 RabbitMQ 连接，确保程序结束时能正确关闭消息队列连接
	defer func() {
		if err := rabbitmq.Close(); err != nil {
			// 若关闭 RabbitMQ 连接失败，记录错误信息并终止程序
			log.Fatalf("failed to close rabbitmq: %v", err)
		}
	}()
	// 启动一个 goroutine 来消费支付完成的消息
	go rabbitmq.ConsumePaymentCompleted()

	// 创建 Redis 客户端实例，用于缓存操作
	redisClient := cache.NewRedisClient(&cfg.Redis)

	// 初始化服务层，传入订单仓库、RabbitMQ 实例和 Redis 客户端
	orderService := service.NewOrderService(orderRepo, rabbitmq, redisClient)

	// 初始化控制器层，传入订单服务实例
	ordercontroller := controller.NewOrderController(orderService)

	// 创建 gRPC 服务器实例，传入配置信息
	grpcServer := server.NewGRPCServer(cfg)
	// 启动 gRPC 服务器，传入订单控制器实例
	if err := grpcServer.Start(ordercontroller); err != nil {
		// 若 gRPC 服务器启动失败，记录错误信息并终止程序
		log.Fatalf("failed to start gRPC Server: %v", err)
	}
	// 延迟关闭 gRPC 服务器，确保程序结束时能正确关闭服务器
	defer grpcServer.Shutdown()

	// 创建一个缓冲通道，用于接收操作系统信号
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT（通常是 Ctrl+C）和 SIGTERM 信号，将信号发送到 quit 通道
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞主线程，等待接收到操作系统信号
	<-quit

	// 当接收到信号后，输出日志信息，表示服务器即将关闭
	log.Println("Shutting down the server...")
}
