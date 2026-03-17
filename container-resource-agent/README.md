# 容器资源管理 Agent

基于 Kubernetes 和 cgroups v2 的智能容器资源管理解决方案。

## 背景

- **平台**: Kubernetes 1.25+, containerd/docker
- **内核**: cgroups v2
- **场景**: 华为云 CCE、自建 K8s 集群
- **解决问题**: 资源争抢、OOM、性能抖动

## 架构概览

```
┌─────────────────────────────────────────────────────────┐
│           Container Resource Management Agent            │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Knowledge    │  │ Diagnostic   │  │ Configuration│  │
│  │ Base         │  │ Engine       │  │ Advisor      │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Monitoring   │  │ Remediation  │  │ Reporting    │  │
│  │ Collector    │  │ Executor     │  │ Generator    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
          │                    │                    │
          ▼                    ▼                    ▼
   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
   │ Kubernetes  │     │ Container   │     │ Node        │
   │ API Server  │     │ Runtime     │     │ System      │
   └─────────────┘     └─────────────┘     └─────────────┘
```

## 核心模块

### 1. 知识库模块 (`knowledge/`)
- cgroups v1/v2 差异和迁移指南
- 资源类型和控制参数
- QoS 类别定义
- 运行时配置参考

### 2. 诊断引擎 (`diagnostic/`)
- 资源使用分析
- 争抢检测
- OOM 诊断
- 碎片分析

### 3. 配置顾问 (`advisor/`)
- Request/Limit 推荐
- LimitRange/ResourceQuota 模板
- 节点预留策略
- QoS 优化建议

### 4. 问题解决器 (`remediation/`)
- Java 应用内存配置
- CPU 密集型任务调度
- 内存泄漏处理
- 多租户隔离

### 5. 监控采集器 (`monitoring/`)
- Prometheus 指标
- cgroup 统计
- 运行时指标
- 内核事件

## 快速开始

```bash
# 部署 Agent
kubectl apply -f deploy/

# 运行诊断
kubectl exec -it container-resource-agent -- /app/diagnose --namespace=default

# 获取建议
kubectl exec -it container-resource-agent -- /app/advise --pod=my-app

# 应用修复
kubectl exec -it container-resource-agent -- /app/remediate --issue=oom
```

## 使用场景

1. **资源规划**: 为新应用推荐合适的资源配置
2. **故障诊断**: 分析 OOM、性能抖动等问题
3. **容量优化**: 识别资源浪费和碎片
4. **安全加固**: 确保多租户资源隔离

## 文档结构

```
container-resource-agent/
├── knowledge/          # 知识库
│   ├── cgroups.md      # cgroups 详解
│   ├── qos.md          # QoS 类别
│   └── runtime.md      # 运行时配置
├── diagnostic/         # 诊断引擎
│   ├── analyzer.go     # 资源分析
│   ├── detector.go     # 问题检测
│   └── fragment.go     # 碎片分析
├── advisor/            # 配置顾问
│   ├── recommender.go  # 资源推荐
│   ├── templates/      # 配置模板
│   └── policies/       # 策略定义
├── remediation/        # 问题解决
│   ├── java.go         # Java 应用
│   ├── cpu.go          # CPU 任务
│   └── isolation.go    # 多租户隔离
├── monitoring/         # 监控采集
│   ├── collector.go    # 指标采集
│   ├── metrics.go      # 指标定义
│   └── alerts/         # 告警规则
└── deploy/             # 部署文件
    ├── agent.yaml      # Agent 部署
    ├── rbac.yaml       # 权限配置
    └── config.yaml     # 配置文件
```
