package main

import (
	"context"
	"log"
	"order-microsystem/payment-service/internal/controller"
	"order-microsystem/payment-service/internal/domain/repository"
	"order-microsystem/payment-service/internal/server"
	"order-microsystem/payment-service/internal/service"
	"order-microsystem/payment-service/pkg/config"
	"order-microsystem/payment-service/pkg/database"
	"order-microsystem/payment-service/pkg/messaging"
	"order-microsystem/payment-service/pkg/tracing"
	"os/signal"
	"syscall"
	"time"
)

// main 函数是程序的入口点，负责初始化系统组件、启动服务并处理优雅关闭逻辑。
func main() {
	// 调用 config.NewConfig 函数加载配置文件，配置文件名为 "config"。
	// 若加载失败，使用 log.Fatalf 输出错误信息并终止程序。
	cfg, err := config.NewConfig("config")
	if err != nil {
		log.Fatalf("failed to load config file: %v", err)
	}

	// 调用 database.InitMySQL 函数初始化 MySQL 数据库连接，传入 MySQL 配置。
	// 若连接失败，使用 log.Fatalf 输出错误信息并终止程序。
	db, err := database.InitMySQL(&cfg.Database.MySQL)
	if err != nil {
		log.Fatalf("failed to connect MySQL: %v", err)
	}

	// 创建 MySQL 仓库实例，用于与数据库进行交互。
	repo := repository.NewMySQLRepository(db)
	// 调用仓库的 AutoMigration 方法自动迁移数据库表结构。
	// 若迁移失败，使用 log.Fatalf 输出错误信息并终止程序。
	if err = repo.AutoMigration(); err != nil {
		log.Fatalf("failed to call AutoMigration: %v", err)
	}

	// 调用 messaging.NewRabbitMQ 函数初始化 RabbitMQ 连接，传入 RabbitMQ 配置和仓库实例。
	// 若连接失败，使用 log.Fatalf 输出错误信息并终止程序。
	rabbitMQ, err := messaging.NewRabbitMQ(&cfg.RabbitMQ, repo)
	if err != nil {
		log.Fatalf("failed to connect RabbitMQ: %v", err)
	}
	// 延迟关闭 RabbitMQ 连接，在函数返回时执行 Close 方法。
	// 注意：这里 Close 方法可能返回错误，需要处理，当前未处理。
	defer rabbitMQ.Close()
	// 启动一个 goroutine 来消费库存锁定消息。
	go rabbitMQ.ConsumeInventoryLocked()

	// 调用 tracing.InitTracer 函数初始化分布式追踪器，传入配置信息。
	// 若初始化失败，使用 log.Fatalf 输出错误信息并终止程序。
	tracerProvider, err := tracing.InitTracer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	// 延迟关闭追踪器，在函数返回时执行 Shutdown 方法，并处理可能的错误。
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Fatalf("error shutting down tracer provider: %v", err)
		}
	}()

	// 创建支付服务实例，传入数据库仓库实例。
	paymentService := service.NewPaymentService(repo)
	// 创建支付控制器实例，传入支付服务实例。
	paymentController := controller.NewPaymentController(paymentService)

	// 创建 gRPC 服务器实例，传入配置信息。
	grpcServer := server.NewGRPCServer(cfg)

	// 创建一个可通知的上下文，监听 SIGINT 和 SIGTERM 信号。
	// 当接收到这些信号时，上下文会被取消。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// 延迟停止信号监听，在函数返回时执行。
	defer stop()

	// 启动一个 goroutine 来启动 gRPC 服务器，传入支付控制器实例。
	// 若启动失败，使用 log.Fatalf 输出错误信息并终止程序。
	go func() {
		if err := grpcServer.Start(paymentController); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// 阻塞等待终止信号，当接收到信号时，ctx.Done() 通道会接收到值。
	<-ctx.Done()
	// 输出日志信息，表示开始优雅关闭服务器。
	log.Println("shutting down server gracefully...")

	// 创建一个带 10 秒超时的上下文，用于控制服务器关闭操作的时间。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 延迟取消上下文，避免资源泄漏。
	defer cancel()

	// 启动一个 goroutine 来关闭 gRPC 服务器。
	go func() {
		grpcServer.Shutdown()
	}()

	// 等待关闭完成或超时。
	// 若超时，输出错误信息并终止程序；若正常关闭，输出成功信息。
	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Fatal("Graceful shutdown timed out")
		} else {
			log.Println("Shutdown completed successfully")
		}
	}
}
