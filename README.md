# 系统架构图

​	本项目是一个高并发、可扩展的分布式系统，采用微服务架构设计，整体系统拆分为API网关服务、订单服务、库存服务和支付服务，各服务独立部署、协同工作，有效提升系统的可维护性、可扩展性。系统通过服务发现，分布式追踪，监控告警等机制，保障了服务的稳定性和可观测性。

- **编程语言**：Go
- **容器化部署**：Docker、Docker Compose
- **服务发现**：Consul
- **服务通信**：gRPC
- **分布式追踪**：OpenTelemetry、Jaeger
- **监控告警**：Prometheus、Grafana、Alertmanager
- **日志采集**：ELK Stack
- **熔断限流**：Hystrix-Go
- **消息队列**：RabbitMQ
- **数据库**：MongoDB、MySQL
- **缓存**：Redis

```MERMAID
sequenceDiagram
    participant User
    participant OrderService
    participant RabbitMQ
    participant InventoryService
    participant PaymentService

    User->>OrderService: 提交订单
    OrderService->>RabbitMQ: 发送"订单创建"事件 (order.created)
    RabbitMQ->>InventoryService: 扣减库存 (库存服务消费消息)
    InventoryService->>RabbitMQ: 发送"库存锁定"事件 (inventory.locked)
    RabbitMQ->>PaymentService: 发起支付 (支付服务消费消息)
    PaymentService->>RabbitMQ: 发送"支付成功"事件 (payment.completed)
    RabbitMQ->>OrderService: 更新订单状态为"已完成"
```



```MERMAID
graph TD
    A[订单创建] --> B[扣减库存]
    B -- 成功 --> C[发起支付]
    B -- 失败 --> D[订单状态: 库存不足]
    C -- 成功 --> E[订单完成]
    C -- 失败 --> F[释放库存]
    F --> G[订单状态: 支付失败]
    H[消息丢失] --> I[定时任务补偿]
```

