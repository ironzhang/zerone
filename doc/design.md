# zerone设计文档

## 目标

用Go实现一个简单易用、高性能的服务治理型的RPC框架。

## 功能性需求分析

- 支持多种编码方式(zerone提供json和protobuf两种编码器的实现)
- 服务发现
	* 静态机制
	* 动态机制(zerone提供etcd服务注册发现的实现)
- 支持自动重连
- 支持随机、轮询、权重、哈希等多种负载均衡策略
- 支持点对点和广播这两种调用策略
- 支持Failover、Failfast、Failtry等多种失败重试策略
- 限流(server和client都要支持)、熔断、超时
- 日志追踪(zerone框架日志、RPC调用链追踪)、指标监控

