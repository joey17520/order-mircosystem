package logger

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"order-microsystem/api-service/pkg/config"
	"os"
	"sync"
	"time"
)

// LogStashHook 定义了一个 Logrus 钩子，用于将日志发送到 Logstash。
type LogStashHook struct {
	conn      net.Conn       // 与 Logstash 建立的网络连接
	hostname  string         // 主机名
	appName   string         // 应用名称
	mux       sync.Mutex     // 互斥锁，用于保证并发安全
	queue     chan []byte    // 日志消息队列，用于异步发送日志
	waitGroup sync.WaitGroup // 等待组，用于等待所有日志消息发送完成
}

// NewLogStashHook 创建一个新的 LogStashHook 实例。
// proto 是网络协议，如 "tcp" 或 "udp"。
// addr 是 LogStash 的地址。
// appName 是应用的名称。
// queueSize 是日志消息队列的大小，若为 0 则不使用队列。
func NewLogStashHook(proto string, addr string, appName string, queueSize int) (*LogStashHook, error) {
	// 建立与 LogStash 的网络连接
	conn, err := net.Dial(proto, addr)
	if err != nil {
		return nil, err
	}

	// 获取当前主机名
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	// 初始化 LogStashHook 实例
	hook := &LogStashHook{
		conn:     conn,
		hostname: hostname,
		appName:  appName,
	}

	// 如果队列大小大于 0，则创建队列并启动处理协程
	if queueSize > 0 {
		hook.queue = make(chan []byte, queueSize)
		go hook.processQueue()
	}

	return hook, nil
}

// processQueue 处理日志消息队列，将队列中的日志消息发送到 LogStash。
func (hook *LogStashHook) processQueue() {
	for logEntry := range hook.queue {
		// 加锁，保证并发安全
		hook.mux.Lock()
		// 将日志消息写入 LogStash 连接
		_, err := hook.conn.Write(logEntry)
		// 解锁
		hook.mux.Unlock()
		// 标记一个任务完成
		hook.waitGroup.Done()

		if err != nil {
			// 若写入失败，打印错误信息
			fmt.Println("LogStash写入错误: ", err.Error())
		}
	}
}

// Fire 当有日志条目需要记录时，该方法会被调用，将日志条目发送到 LogStash。
func (hook *LogStashHook) Fire(entry *logrus.Entry) error {
	// 构建日志数据
	data := map[string]interface{}{
		"@timestamp": time.Now().Format(time.RFC3339Nano), // 日志时间戳
		"message":    entry.Message,                       // 日志消息
		"level":      entry.Level.String(),                // 日志级别
		"hostname":   hook.hostname,                       // 主机名
		"app_name":   hook.appName,                        // 应用名称
		"service":    hook.appName,                        // 服务名称
	}

	// 将日志条目的额外数据添加到日志数据中
	for k, v := range entry.Data {
		data[k] = v
	}

	// 将日志数据转换为 JSON 格式
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// 添加换行符
	jsonData = append(jsonData, '\n')

	// 如果使用队列，则将日志消息放入队列
	if hook.queue != nil {
		hook.waitGroup.Add(1)
		hook.queue <- jsonData
	} else {
		// 若不使用队列，直接将日志消息写入 LogStash 连接
		hook.mux.Lock()
		_, err := hook.conn.Write(jsonData)
		hook.mux.Unlock()
		return err
	}

	return nil
}

// Levels 返回该钩子支持的日志级别。
func (hook *LogStashHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

// Close 关闭 LogStash 连接，等待所有日志消息发送完成。
func (hook *LogStashHook) Close() {
	// 若使用队列，等待所有日志消息发送完成
	if hook.queue != nil {
		hook.waitGroup.Wait()
	}
	// 关闭网络连接
	hook.conn.Close()
}

// InitLogger 初始化一个 Logrus 日志实例，并根据配置添加 LogStash 钩子。
// appName 是应用的名称。
// logStashAddr 是 LogStash 的地址，若为空则不添加 LogStash 钩子。
// async 表示是否使用异步模式发送日志。
func InitLogger(config *config.LoggerConfig) *logrus.Logger {
	logStashAddr := fmt.Sprintf("%s:%d", config.LogStashHost, config.LogStashPort)
	// 创建一个新的 Logrus 日志实例
	log := logrus.New()
	// 设置日志格式为 JSON 格式，并指定时间戳格式
	log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})

	// 如果 LogStash 地址不为空，则添加 LogStash 钩子
	if logStashAddr != "" {
		var queueSize int
		// 若为异步模式，设置队列大小
		if config.Async {
			queueSize = 1000
		}
		// 创建 LogStashHook 实例
		hook, err := NewLogStashHook("tcp", logStashAddr, config.ServiceName, queueSize)
		if err != nil {
			// 若创建失败，打印错误信息
			fmt.Println("初始化LogStashHook失败: ", err.Error())
		}
		// 将 LogStashHook 添加到日志实例中
		log.AddHook(hook)
	}

	return log
}
