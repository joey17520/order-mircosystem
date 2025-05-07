package main

import (
	"log"
	"order-microsystem/api-service/internal/server"
	"order-microsystem/api-service/pkg/config"
	"order-microsystem/api-service/pkg/logger"
	"order-microsystem/api-service/pkg/tracing"
	"os"
	"os/signal"
	"syscall"
)

// main 是程序的入口函数，负责初始化配置、追踪器，启动 HTTP 服务器，
// 并处理优雅关闭信号。
func main() {
	// 调用 config.LoadConfig 函数加载配置文件，配置文件名为 "config"。
	// 如果加载失败，使用 log.Fatal 输出错误信息并终止程序。
	cfg, err := config.LoadConfig("config")
	if err != nil {
		// 原代码此处存在格式错误，%v 不会被正确替换，应使用 log.Fatalf
		log.Fatalf("Failed to Read config file: %v", err)
	}

	// 调用 tracing.InitTracer 函数初始化追踪器，传入加载好的配置。
	// 如果初始化失败，使用 log.Fatalf 输出错误信息并终止程序。
	tp, err := tracing.InitTracer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}

	// 初始化日志
	logStash := logger.InitLogger(&cfg.Logger)

	// 调用 server.NewHTTPServer 函数创建一个新的 HTTP 服务器实例，
	// 传入配置和追踪器。
	httpServer := server.NewHTTPServer(cfg, tp, logStash)
	// 使用 goroutine 异步启动 HTTP 服务器，避免阻塞主线程。
	// 如果启动失败，使用 log.Fatalf 输出错误信息并终止程序。
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	// 创建一个缓冲通道 quit，用于接收操作系统信号。
	quit := make(chan os.Signal, 1)
	// 使用 signal.Notify 函数监听 SIGINT（通常是 Ctrl+C）和 SIGTERM 信号，
	// 当接收到这些信号时，将信号发送到 quit 通道。
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞主线程，等待接收到操作系统信号。
	<-quit

	// 当接收到信号后，输出日志信息，表示服务器即将关闭。
	log.Println("shutting down server...")
}
