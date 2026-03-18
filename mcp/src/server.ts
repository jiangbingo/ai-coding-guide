/**
 * MCP Server for Huawei Cloud Kernel Tools
 * 
 * 提供 KVM 优化、性能分析、内核调试等工具
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
  ListResourcesRequestSchema,
  ReadResourceRequestSchema,
  ListPromptsRequestSchema,
  GetPromptRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';

// 创建 MCP 服务器
const server = new Server({
  name: 'huawei-cloud-kernel-tools',
  version: '1.0.0',
}, {
  capabilities: {
    tools: {},
    resources: {},
    prompts: {},
  },
});

// ============================================
// Tools - 工具定义
// ============================================

server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: [
      {
        name: 'kvm_optimize',
        description: '优化 KVM 虚拟机性能，提供 CPU、内存、网络、存储等方面的配置建议',
        inputSchema: {
          type: 'object',
          properties: {
            vm_type: {
              type: 'string',
              enum: ['kvm', 'qemu', 'xen'],
              description: '虚拟机类型',
            },
            workload: {
              type: 'string',
              enum: ['cpu-intensive', 'memory-intensive', 'io-intensive', 'balanced'],
              description: '工作负载类型',
            },
            current_config: {
              type: 'object',
              properties: {
                vcpus: { type: 'number' },
                memory_gb: { type: 'number' },
                network_type: { type: 'string' },
                storage_type: { type: 'string' },
              },
            },
          },
          required: ['vm_type'],
        },
      },
      {
        name: 'perf_analyze',
        description: '分析系统性能瓶颈，识别 CPU、内存、I/O 等方面的性能问题',
        inputSchema: {
          type: 'object',
          properties: {
            metrics: {
              type: 'object',
              properties: {
                cpu_usage: { type: 'number' },
                memory_usage: { type: 'number' },
                disk_io: { type: 'number' },
                network_io: { type: 'number' },
              },
            },
            duration_seconds: {
              type: 'number',
              description: '性能数据采集时长（秒）',
            },
          },
          required: ['metrics'],
        },
      },
      {
        name: 'kernel_debug',
        description: '内核调试助手，提供内核崩溃、死锁、内存泄漏等问题的诊断建议',
        inputSchema: {
          type: 'object',
          properties: {
            issue_type: {
              type: 'string',
              enum: ['crash', 'hang', 'memory_leak', 'performance', 'unknown'],
              description: '问题类型',
            },
            logs: {
              type: 'string',
              description: '内核日志或错误信息',
            },
            kernel_version: {
              type: 'string',
              description: '内核版本',
            },
          },
          required: ['issue_type'],
        },
      },
      {
        name: 'numa_analyze',
        description: 'NUMA 亲和性分析，优化内存分配和 CPU 绑定策略',
        inputSchema: {
          type: 'object',
          properties: {
            numa_nodes: {
              type: 'number',
              description: 'NUMA 节点数量',
            },
            workload_type: {
              type: 'string',
              enum: ['single-threaded', 'multi-threaded', 'database', 'hpc'],
              description: '工作负载类型',
            },
          },
          required: ['numa_nodes'],
        },
      },
      {
        name: 'virtio_configure',
        description: 'VirtIO 设备配置优化，提升虚拟化 I/O 性能',
        inputSchema: {
          type: 'object',
          properties: {
            device_type: {
              type: 'string',
              enum: ['net', 'blk', 'scsi', 'balloon'],
              description: 'VirtIO 设备类型',
            },
            features: {
              type: 'array',
              items: { type: 'string' },
              description: '要启用的特性列表',
            },
          },
          required: ['device_type'],
        },
      },
    ],
  };
});

// ============================================
// Tool Execution - 工具执行
// ============================================

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;
  
  try {
    switch (name) {
      case 'kvm_optimize':
        return await handleKVMOptimize(args);
      case 'perf_analyze':
        return await handlePerfAnalyze(args);
      case 'kernel_debug':
        return await handleKernelDebug(args);
      case 'numa_analyze':
        return await handleNUMAAnalyze(args);
      case 'virtio_configure':
        return await handleVirtIOConfigure(args);
      default:
        throw new Error(`Unknown tool: ${name}`);
    }
  } catch (error) {
    return {
      content: [{
        type: 'text',
        text: `Error: ${error.message}`,
      }],
      isError: true,
    };
  }
});

// ============================================
// Tool Handlers - 工具处理器
// ============================================

async function handleKVMOptimize(args) {
  const { vm_type, workload, current_config } = args;
  const recommendations = [];
  
  // CPU 优化
  if (workload === 'cpu-intensive') {
    recommendations.push({
      category: 'CPU',
      priority: 'HIGH',
      items: [
        '启用 vCPU 绑定（CPU pinning）',
        '使用 host-passthrough 模式',
        '禁用 CPU 超额订阅',
        '配置 CPU 独占模式',
      ],
    });
  }
  
  // 内存优化
  recommendations.push({
    category: 'Memory',
    priority: 'HIGH',
    items: [
      '启用大页内存（HugePages）',
      '配置 NUMA 亲和性',
      '禁用 KSM（内存合并）',
      '使用 virtio-balloon 动态调整',
    ],
  });
  
  // 网络 I/O 优化
  if (workload === 'io-intensive') {
    recommendations.push({
      category: 'Network I/O',
      priority: 'HIGH',
      items: [
        '使用 virtio-net + vhost-net',
        '启用多队列（multi-queue）',
        '配置 RSS（Receive Side Scaling）',
        '使用 SR-IOV 直通（如可用）',
      ],
    });
  }
  
  // 存储 I/O 优化
  recommendations.push({
    category: 'Storage I/O',
    priority: 'MEDIUM',
    items: [
      '使用 virtio-blk 或 virtio-scsi',
      '启用 IOThread',
      '配置 native AIO',
      '使用 NVMe 直通（如可用）',
    ],
  });
  
  return {
    content: [{
      type: 'text',
      text: JSON.stringify({
        summary: `KVM ${vm_type} 优化建议`,
        workload: workload || 'balanced',
        current_config,
        recommendations,
        generatedAt: new Date().toISOString(),
      }, null, 2),
    }],
  };
}

async function handlePerfAnalyze(args) {
  const { metrics, duration_seconds } = args;
  const analysis = {
    summary: '性能分析报告',
    duration: duration_seconds || 60,
    bottlenecks: [],
    recommendations: [],
  };
  
  // CPU 瓶颈检测
  if (metrics.cpu_usage > 80) {
    analysis.bottlenecks.push({
      type: 'CPU',
      severity: 'HIGH',
      value: metrics.cpu_usage,
      description: 'CPU 使用率过高，可能影响系统响应',
    });
    analysis.recommendations.push('检查是否有 CPU 密集型进程');
    analysis.recommendations.push('考虑 CPU 绑定或调度优化');
  }
  
  // 内存瓶颈检测
  if (metrics.memory_usage > 90) {
    analysis.bottlenecks.push({
      type: 'Memory',
      severity: 'HIGH',
      value: metrics.memory_usage,
      description: '内存使用率接近上限，可能触发 OOM',
    });
    analysis.recommendations.push('检查内存泄漏');
    analysis.recommendations.push('调整大页内存配置');
  }
  
  // I/O 瓶颈检测
  if (metrics.disk_io > 80) {
    analysis.bottlenecks.push({
      type: 'Disk I/O',
      severity: 'MEDIUM',
      value: metrics.disk_io,
      description: '磁盘 I/O 压力较大',
    });
    analysis.recommendations.push('检查 I/O 调度器配置');
    analysis.recommendations.push('考虑使用更快的存储设备');
  }
  
  return {
    content: [{
      type: 'text',
      text: JSON.stringify(analysis, null, 2),
    }],
  };
}

async function handleKernelDebug(args) {
  const { issue_type, logs, kernel_version } = args;
  
  const diagnosis = {
    issue_type,
    kernel_version: kernel_version || 'unknown',
    analysis: [],
    next_steps: [],
  };
  
  switch (issue_type) {
    case 'crash':
      diagnosis.analysis.push('检测到内核崩溃');
      diagnosis.next_steps.push('分析 dmesg 日志');
      diagnosis.next_steps.push('检查 /proc/sys/kernel/tainted');
      diagnosis.next_steps.push('使用 crash 工具分析 vmcore');
      break;
    case 'hang':
      diagnosis.analysis.push('检测到系统挂起');
      diagnosis.next_steps.push('检查是否为死锁');
      diagnosis.next_steps.push('使用 SysRq 收集诊断信息');
      diagnosis.next_steps.push('分析 /proc/locks');
      break;
    case 'memory_leak':
      diagnosis.analysis.push('检测到可能的内存泄漏');
      diagnosis.next_steps.push('使用 kmemleak 检测');
      diagnosis.next_steps.push('分析 /proc/meminfo');
      diagnosis.next_steps.push('使用 BPF 工具追踪内存分配');
      break;
    default:
      diagnosis.analysis.push('问题类型未知，需要更多信息');
      diagnosis.next_steps.push('收集完整的系统日志');
      diagnosis.next_steps.push('使用 perf/ebpf 进行性能分析');
  }
  
  return {
    content: [{
      type: 'text',
      text: JSON.stringify(diagnosis, null, 2),
    }],
  };
}

async function handleNUMAAnalyze(args) {
  const { numa_nodes, workload_type } = args;
  
  const analysis = {
    numa_nodes,
    workload_type: workload_type || 'multi-threaded',
    recommendations: [],
    configuration: {},
  };
  
  if (numa_nodes > 1) {
    analysis.recommendations.push('启用 NUMA 感知的内存分配');
    analysis.recommendations.push('使用 numactl 绑定进程到特定节点');
    
    if (workload_type === 'database') {
      analysis.configuration = {
        memory_policy: 'localalloc',
        cpu_affinity: 'node-wise',
        hugepages_per_node: 1024,
      };
    } else if (workload_type === 'hpc') {
      analysis.configuration = {
        memory_policy: 'interleave',
        cpu_affinity: 'spread',
        hugepages_per_node: 2048,
      };
    }
  } else {
    analysis.recommendations.push('单 NUMA 节点，无需特殊优化');
  }
  
  return {
    content: [{
      type: 'text',
      text: JSON.stringify(analysis, null, 2),
    }],
  };
}

async function handleVirtIOConfigure(args) {
  const { device_type, features } = args;
  
  const config = {
    device_type,
    features: features || [],
    recommended_features: [],
    configuration: {},
  };
  
  switch (device_type) {
    case 'net':
      config.recommended_features = [
        'csum', 'gso', 'tso', 'ufo', 'mergeable_rx_bufs',
        'status', 'mrg_rxbuf', 'ctrl_vq', 'ctrl_rx', 'ctrl_vlan',
        'guest_tso4', 'guest_tso6', 'guest_ecn', 'guest_ufo',
      ];
      config.configuration = {
        queues: 4,
        rx_queue_size: 1024,
        tx_queue_size: 1024,
      };
      break;
    case 'blk':
      config.recommended_features = [
        'size_max', 'seg_max', 'geometry', 'ro', 'blk_size',
        'toascii', 'flush', 'topology', 'config_wce', 'discard',
      ];
      config.configuration = {
        iothread: 'iothread1',
        queue_size: 256,
        native_aio: true,
      };
      break;
    case 'scsi':
      config.recommended_features = [
        'inout', 'any_layout', 'indirect_desc', 'event_idx',
      ];
      config.configuration = {
        num_queues: 4,
        queue_size: 128,
        iothread: 'iothread1',
      };
      break;
  }
  
  return {
    content: [{
      type: 'text',
      text: JSON.stringify(config, null, 2),
    }],
  };
}

// ============================================
// Resources - 资源定义
// ============================================

server.setRequestHandler(ListResourcesRequestSchema, async () => {
  return {
    resources: [
      {
        uri: 'kernel://docs',
        name: 'Linux Kernel Documentation',
        description: 'Linux 内核官方文档',
        mimeType: 'text/markdown',
      },
      {
        uri: 'kvm://best-practices',
        name: 'KVM Best Practices Guide',
        description: 'KVM 虚拟化最佳实践指南',
        mimeType: 'text/markdown',
      },
      {
        uri: 'perf://tools',
        name: 'Performance Analysis Tools Reference',
        description: '性能分析工具参考手册',
        mimeType: 'text/markdown',
      },
    ],
  };
});

server.setRequestHandler(ReadResourceRequestSchema, async (request) => {
  const { uri } = request.params;
  
  // 这里应该从实际资源加载内容
  // 目前返回示例内容
  const resources = {
    'kernel://docs': '# Linux Kernel Documentation\n\n...',
    'kvm://best-practices': '# KVM Best Practices\n\n...',
    'perf://tools': '# Performance Analysis Tools\n\n...',
  };
  
  const content = resources[uri] || 'Resource not found';
  
  return {
    contents: [{
      uri,
      mimeType: 'text/markdown',
      text: content,
    }],
  };
});

// ============================================
// Prompts - 提示词定义
// ============================================

server.setRequestHandler(ListPromptsRequestSchema, async () => {
  return {
    prompts: [
      {
        name: 'kvm-tuning',
        description: 'KVM 虚拟机性能调优提示',
        arguments: [
          {
            name: 'workload_type',
            description: '工作负载类型（cpu-intensive/memory-intensive/io-intensive）',
            required: true,
          },
        ],
      },
      {
        name: 'kernel-debug-workflow',
        description: '内核调试工作流提示',
        arguments: [
          {
            name: 'issue_type',
            description: '问题类型（crash/hang/memory_leak）',
            required: true,
          },
        ],
      },
    ],
  };
});

server.setRequestHandler(GetPromptRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;
  
  if (name === 'kvm-tuning') {
    return {
      description: `KVM ${args.workload_type} 工作负载调优`,
      messages: [
        {
          role: 'user',
          content: {
            type: 'text',
            text: `请针对 ${args.workload_type} 工作负载，提供 KVM 虚拟机性能调优建议。

重点关注：
1. CPU 调度和绑定策略
2. 内存分配和 NUMA 亲和性
3. 网络 I/O 优化（virtio 配置）
4. 存储 I/O 优化（磁盘调度器）

请提供具体的配置示例和验证方法。`,
          },
        },
      ],
    };
  }
  
  throw new Error(`Unknown prompt: ${name}`);
});

// ============================================
// Start Server - 启动服务器
// ============================================

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error('Huawei Cloud Kernel Tools MCP server running on stdio');
}

main().catch((error) => {
  console.error('Fatal error in main():', error);
  process.exit(1);
});