package main

import (
	"context"
	"log"
	"order-microsystem/inventory-service/internal/controller"
	"order-microsystem/inventory-service/internal/domain/repository"
	"order-microsystem/inventory-service/internal/server"
	"order-microsystem/inventory-service/internal/service"
	"order-microsystem/inventory-service/pkg/config"
	"order-microsystem/inventory-service/pkg/database"
	"order-microsystem/inventory-service/pkg/messaging"
	"order-microsystem/inventory-service/pkg/tracing"
	"os/signal"
	"syscall"
	"time"
)

// main 是程序的入口函数，负责初始化配置、数据库、消息队列、追踪器等组件，
// 启动 gRPC 服务器，并处理优雅关闭逻辑。
func main() {
	// 加载配置文件，配置文件名为 "config"
	cfg, err := config.NewConfig("config")
	if err != nil {
		// 若配置文件加载失败，记录错误信息并终止程序
		log.Fatalf("failed to load config file: %v", err)
	}

	// 初始化 MySQL 数据库连接，传入 MySQL 配置信息
	db, err := database.InitMySQL(&cfg.Database.MySQL)
	if err != nil {
		// 若数据库连接失败，记录错误信息并终止程序
		log.Fatalf("failed to connect MySQL: %v", err)
	}

	// 创建 MySQL 仓库实例，用于操作数据库
	repo := repository.NewMySQLRepository(db)
	// 执行数据库自动迁移，创建或更新数据库表结构
	if err = repo.AutoMigrations(); err != nil {
		// 若自动迁移失败，记录错误信息并终止程序
		log.Fatalf("failed to call AutoMigration: %v", err)
	}

	// 初始化 RabbitMQ 连接，传入 RabbitMQ 配置信息和数据库仓库实例
	rabbitMQ, err := messaging.NewRabbitMQ(&cfg.RabbitMQ, repo)
	if err != nil {
		// 若 RabbitMQ 连接失败，记录错误信息并终止程序
		log.Fatalf("failed to connect RabbitMQ: %v", err)
	}
	// 延迟关闭 RabbitMQ 连接，在函数返回时执行
	defer rabbitMQ.Close()
	// 启动一个 goroutine 来消费 RabbitMQ 中订单创建的消息
	go rabbitMQ.ConsumeOrderCreated()

	// 初始化分布式追踪器，传入配置信息
	tracerProvider, err := tracing.InitTracer(cfg)
	if err != nil {
		// 若追踪器初始化失败，记录错误信息并终止程序
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	// 延迟关闭追踪器，在函数返回时执行，并处理关闭过程中可能出现的错误
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Fatalf("error shutting down tracer provider: %v", err)
		}
	}()

	// 创建库存服务实例，传入数据库仓库实例
	inventoryService := service.NewInventoryService(repo)
	// 创建库存控制器实例，传入库存服务实例
	inventoryController := controller.NewInventoryController(inventoryService)

	// 创建 gRPC 服务器实例，传入配置信息
	grpcServer := server.NewGRPCServer(cfg)

	// 创建一个可通知的上下文，监听 SIGINT 和 SIGTERM 信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// 延迟停止信号监听，在函数返回时执行
	defer stop()

	// 启动一个 goroutine 来启动 gRPC 服务器
	go func() {
		if err := grpcServer.Start(inventoryController); err != nil {
			// 若 gRPC 服务器启动失败，记录错误信息并终止程序
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// 等待终止信号，当接收到信号时，ctx.Done() 通道会接收到值
	<-ctx.Done()
	// 记录日志，表示开始优雅关闭服务器
	log.Println("shutting down server gracefully...")

	// 创建一个带 10 秒超时的上下文，用于控制关闭操作的时间
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 延迟取消上下文，避免资源泄漏
	defer cancel()

	// 启动一个 goroutine 来关闭 gRPC 服务器
	go func() {
		grpcServer.Shutdown()
	}()

	// 等待关闭完成或超时
	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			// 若关闭超时，记录错误信息并终止程序
			log.Fatal("Graceful shutdown timed out")
		} else {
			// 若关闭成功，记录日志信息
			log.Println("Shutdown completed successfully")
		}
	}
}
