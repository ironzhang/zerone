# 开发计划

## TODO

### 文档

- [ ] 注释
- [ ] 示例
- [ ] README.md
- [ ] Design.md

### 功能特性

- [ ] 监控指标: 如调用次数等
- [ ] 限流、熔断(低优先级)
- [ ] 支持protobuf编解码协议

### 功能优化

- [ ] 梳理ServerName, ClientName, ServiceName, ClassName, MethodName等概念和从属关系，修改代码中的命名
- [ ] 抽取trace output event对象，输出事件给外部接口
- [ ] trace日志输出更多详细信息: 如IP等
