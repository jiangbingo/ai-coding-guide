# 容器运行时资源管理

## 1. containerd 配置

### 完整配置文件

```toml
# /etc/containerd/config.toml
version = 2

[plugins."io.containerd.grpc.v1.cri"]
  # Cgroup 配置
  disable_cgroup = false
  systemd_cgroup = true  # cgroups v2 推荐

  # 沙箱配置
  sandbox_image = "k8s.gcr.io/pause:3.9"

  # 容器统计收集间隔
  stats_collect_period = 10

  [plugins."io.containerd.grpc.v1.cri".containerd]
    snapshotter = "overlayfs"
    default_runtime_name = "runc"

    # 运行时配置
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
      runtime_type = "io.containerd.runc.v2"
      runtime_engine = ""
      runtime_root = ""

      # cgroups v2 配置
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
        SystemdCgroup = true
        # 二进制路径
        BinaryName = "runc"

    # 性能调优
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
      NoPivotRoot = false
      NoNewKeyring = false
      ShimCgroup = ""
      IoUid = 0
      IoGid = 0

  # CNI 配置
  [plugins."io.containerd.grpc.v1.cri".cni]
    bin_dir = "/opt/cni/bin"
    conf_dir = "/etc/cni/net.d"
    max_conf_num = 1

  # 镜像配置
  [plugins."io.containerd.grpc.v1.cri".registry]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
        endpoint = ["https://registry-1.docker.io"]

  # 资源限制默认值
  [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime]
    runtime_type = "io.containerd.runc.v2"
    [plugins."io.containerd.grpc.v1.cri".containerd.default_runtime.options]
      SystemdCgroup = true
```

### 资源管理相关配置

```toml
[plugins."io.containerd.grpc.v1.cri"]
  # 容器停止超时
  stop_timeout = "30s"

  # 镜像拉取超时
  image_pull_progress_timeout = "5m0s"

  # 最大并发镜像拉取
  max_concurrent_downloads = 3

  # 镜像拉取重试
  image_pull_with_sync_fs = false

  # 磁盘使用限制
  [plugins."io.containerd.grpc.v1.cri".containerd]
    # 清理未使用镜像
    discard_unpacked_layers = false
    # 默认快照器
    snapshotter = "overlayfs"
```

### 华为云 CCE 优化配置

```toml
# 华为云 CCE containerd 优化配置
version = 2

[plugins."io.containerd.grpc.v1.cri"]
  systemd_cgroup = true

  [plugins."io.containerd.grpc.v1.cri".containerd]
    snapshotter = "overlayfs"

    # 针对 Huawei Cloud 优化
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
      runtime_type = "io.containerd.runc.v2"

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
        SystemdCgroup = true
        # 性能优化
        NoPivotRoot = false

  # 华为云镜像加速
  [plugins."io.containerd.grpc.v1.cri".registry]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."swr.cn-north-4.myhuaweicloud.com"]
        endpoint = ["https://swr.cn-north-4.myhuaweicloud.com"]

  # 流量控制
  [plugins."io.containerd.grpc.v1.cri".containerd]
    # 最大并发下载
    max_concurrent_downloads = 5
    # 解压并发
    max_concurrent_uploads = 5
```

## 2. Docker 配置

### daemon.json

```json
{
  "exec-opts": [
    "native.cgroupdriver=systemd"
  ],
  "cgroup-parent": "kubepods",
  "live-restore": true,
  "userland-proxy": false,
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ],
  "default-ulimits": {
    "nofile": {
      "Name": "nofile",
      "Hard": 65535,
      "Soft": 65535
    }
  },
  "default-runtime": "runc",
  "runtimes": {
    "runc": {
      "path": "runc"
    }
  },
  "registry-mirrors": [
    "https://mirror.ccs.tencentyun.com"
  ],
  "insecure-registries": [],
  "debug": false,
  "hosts": ["unix:///var/run/docker.sock"]
}
```

### Docker Compose 资源配置

```yaml
version: '3.8'

services:
  app:
    image: my-app:latest
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
      restart_policy:
        condition: on-failure
        max_attempts: 3
    ulimits:
      nofile:
        soft: 65535
        hard: 65535
```

## 3. runc 配置

### config.json (OCI Runtime Spec)

```json
{
  "ociVersion": "1.0.2",
  "process": {
    "terminal": false,
    "user": {
      "uid": 0,
      "gid": 0
    },
    "args": ["sh"],
    "env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
      "TERM=xterm"
    ],
    "cwd": "/",
    "capabilities": {
      "bounding": ["CAP_NET_BIND_SERVICE"],
      "effective": ["CAP_NET_BIND_SERVICE"],
      "permitted": ["CAP_NET_BIND_SERVICE"]
    },
    "rlimits": [
      {
        "type": "RLIMIT_NOFILE",
        "hard": 65535,
        "soft": 65535
      }
    ],
    "noNewPrivileges": true
  },
  "root": {
    "path": "rootfs",
    "readonly": false
  },
  "hostname": "container",
  "linux": {
    "uidMappings": [],
    "gidMappings": [],
    "resources": {
      "cpu": {
        "shares": 1024,
        "quota": 100000,
        "period": 100000,
        "realtimeRuntime": 0,
        "realtimePeriod": 0,
        "cpus": "0-3",
        "mems": "0"
      },
      "memory": {
        "limit": 1073741824,
        "reservation": 536870912,
        "swap": 2147483648,
        "kernel": 0,
        "kernelTCP": 0,
        "swappiness": 0,
        "disableOOMKiller": false
      },
      "devices": [
        {
          "allow": false,
          "access": "rwm"
        }
      ],
      "blockIO": {
        "weight": 500,
        "leafWeight": 500,
        "throttleReadBpsDevice": [
          {
            "major": 8,
            "minor": 0,
            "rate": 104857600
          }
        ],
        "throttleWriteBpsDevice": [
          {
            "major": 8,
            "minor": 0,
            "rate": 104857600
          }
        ]
      },
      "hugepageLimits": [],
      "network": {
        "classID": 0,
        "priorities": []
      },
      "pids": {
        "limit": 1024
      }
    },
    "cgroupsPath": "/kubepods/burstable/pod123",
    "namespaces": [
      {"type": "pid"},
      {"type": "network"},
      {"type": "ipc"},
      {"type": "uts"},
      {"type": "mount"}
    ]
  }
}
```

## 4. 资源隔离技术

### Namespace 隔离

| Namespace | 隔离内容 | 用途 |
|-----------|----------|------|
| PID | 进程 ID | 容器内进程独立 |
| NET | 网络栈 | 独立网络配置 |
| IPC | System V IPC, POSIX 消息队列 | 进程间通信隔离 |
| UTS | 主机名和域名 | 独立 hostname |
| MNT | 挂载点 | 文件系统隔离 |
| USER | 用户和用户组 ID | 用户权限隔离 |
| CGROUP | Cgroup 根目录 | 资源限制隔离 |

### Capabilities 控制

```yaml
# Kubernetes Pod 配置
apiVersion: v1
kind: Pod
metadata:
  name: restricted-pod
spec:
  containers:
  - name: app
    image: nginx
    securityContext:
      # 丢弃所有能力
      capabilities:
        drop:
        - ALL
        # 只添加必需的能力
        add:
        - NET_BIND_SERVICE
      # 只读根文件系统
      readOnlyRootFilesystem: true
      # 非特权容器
      privileged: false
      # 禁止获取新权限
      allowPrivilegeEscalation: false
```

### Seccomp 配置

```json
{
  "defaultAction": "SCMP_ACT_ERRNO",
  "architectures": ["SCMP_ARCH_X86_64"],
  "syscalls": [
    {
      "names": [
        "accept",
        "accept4",
        "access",
        "arch_prctl",
        "bind",
        "brk",
        "capget",
        "capset",
        "chdir",
        "clock_gettime",
        "close",
        "connect",
        "dup",
        "dup2",
        "dup3",
        "epoll_create",
        "epoll_create1",
        "epoll_ctl",
        "epoll_wait",
        "eventfd",
        "eventfd2",
        "execve",
        "exit",
        "exit_group",
        "fcntl",
        "fstat",
        "futex",
        "getcwd",
        "getdents",
        "getdents64",
        "getegid",
        "geteuid",
        "getgid",
        "getpid",
        "getppid",
        "getrlimit",
        "getsockname",
        "getsockopt",
        "gettid",
        "getuid",
        "ioctl",
        "listen",
        "lseek",
        "madvise",
        "mmap",
        "mprotect",
        "munmap",
        "nanosleep",
        "open",
        "openat",
        "pipe",
        "pipe2",
        "poll",
        "read",
        "readv",
        "recvfrom",
        "recvmsg",
        "rt_sigaction",
        "rt_sigprocmask",
        "rt_sigreturn",
        "sched_getaffinity",
        "sched_yield",
        "sendmsg",
        "sendto",
        "set_robust_list",
        "set_tid_address",
        "setitimer",
        "setsockopt",
        "shutdown",
        "sigaltstack",
        "socket",
        "stat",
        "statfs",
        "sysinfo",
        "uname",
        "wait4",
        "write",
        "writev"
      ],
      "action": "SCMP_ACT_ALLOW"
    }
  ]
}
```

## 5. 性能调优

### 内核参数优化

```bash
# /etc/sysctl.d/99-container.conf

# 网络优化
net.core.somaxconn = 32768
net.ipv4.tcp_max_syn_backlog = 65536
net.core.netdev_max_backlog = 16384
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_tw_reuse = 1
net.ipv4.ip_local_port_range = 1024 65535

# 内存优化
vm.max_map_count = 262144
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
vm.overcommit_memory = 1

# 文件描述符
fs.file-max = 2097152
fs.nr_open = 2097152

# PID 限制
kernel.pid_max = 4194303

# 信号量
kernel.sem = 250 32000 100 128

# 共享内存
kernel.shmmax = 68719476736
kernel.shmall = 4294967296

# cgroups v2
kernel.sched_rt_runtime_us = -1
```

### systemd 配置

```ini
# /etc/systemd/system/kubelet.service.d/10-resources.conf
[Service]
LimitNOFILE=1048576
LimitNPROC=unlimited
LimitCORE=infinity
LimitMEMLOCK=infinity
TasksMax=infinity
CPUAccounting=true
MemoryAccounting=true
```

### containerd 性能调优

```toml
# /etc/containerd/config.toml
version = 2

[debug]
  level = "info"
  format = "json"

[plugins."io.containerd.grpc.v1.cri"]
  # 增加流式缓冲
  stream_server_address = "127.0.0.1"
  stream_server_port = "0"
  stream_idle_timeout = "4h0m0s"

  # 优化镜像拉取
  [plugins."io.containerd.grpc.v1.cri".registry]
    config_path = "/etc/containerd/certs.d"

  # 优化垃圾回收
  [plugins."io.containerd.gc.v1.scheduler"]
    pause_threshold = 0.02
    deletion_threshold = 0
    mutation_threshold = 100
    schedule_delay = "0"
    startup_delay = "100ms"
```

## 6. 故障排查

### 检查运行时状态

```bash
# containerd
ctr --address /run/containerd/containerd.sock version
ctr --address /run/containerd/containerd.sock plugins ls

# Docker
docker info
docker system df
docker system events

# runc
runc --version
runc list
runc state <container-id>
```

### 查看容器资源使用

```bash
# containerd
ctr task metrics <container-id>

# Docker
docker stats --no-stream

# 直接读取 cgroups
cat /sys/fs/cgroup/kubepods/burstable/pod<uid>/memory.current
cat /sys/fs/cgroup/kubepods/burstable/pod<uid>/cpu.stat
```

### 日志分析

```bash
# containerd 日志
journalctl -u containerd -f

# kubelet 日志
journalctl -u kubelet -f

# 内核日志
dmesg -w | grep -i cgroup
dmesg -w | grep -i oom

# 容器日志
crictl logs <container-id>
crictl logs --previous <container-id>
```

### 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| `failed to create containerd task: OCI runtime create failed` | 运行时配置错误 | 检查 runc 版本和配置 |
| `cgroup: cgroup-mountpoint does not exist` | cgroups 未挂载 | 检查 `/sys/fs/cgroup` |
| `failed to write to memory.max: device or resource busy` | cgroups v1/v2 混用 | 统一使用 v2 |
| `permission denied on cgroup` | 权限不足 | 检查 SELinux/AppArmor |
| `container init caused: write /proc/self/attr/keycreate: permission denied` | SELinux 阻止 | 调整 SELinux 策略 |

## 7. 安全加固

### Pod Security Policy (PSP) / Pod Security Standards (PSS)

```yaml
# Pod Security Standards - Restricted
apiVersion: v1
kind: Pod
metadata:
  name: secure-pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    runAsGroup: 1000
    fsGroup: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: app
    image: nginx:alpine
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
    volumeMounts:
    - name: tmp
      mountPath: /tmp
    - name: cache
      mountPath: /var/cache/nginx
    - name: run
      mountPath: /var/run
  volumes:
  - name: tmp
    emptyDir: {}
  - name: cache
    emptyDir: {}
  - name: run
    emptyDir: {}
```

### AppArmor 配置

```bash
# /etc/apparmor.d/containers/container-profile
#include <tunables/global>

profile container-profile flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>

  network inet tcp,
  network inet udp,

  /bin/** ixr,
  /usr/bin/** ixr,
  /lib/** r,
  /lib64/** r,
  /usr/lib/** r,
  /etc/ld.so.* r,

  deny /proc/*/kcore rwxlx,
  deny /proc/*/mem rwxlx,
  deny /proc/** w,

  /proc/** r,
  /sys/** r,

  capability net_bind_service,
}
```

### SELinux 配置

```bash
# 为容器设置 SELinux 上下文
semanage fcontext -a -t container_file_t "/var/lib/containerd(/.*)?"
restorecon -Rv /var/lib/containerd

# 自定义策略
cat > my_container.te <<EOF
policy_module(my_container, 1.0)

require {
  type container_t;
  type container_file_t;
}

allow container_t container_file_t:file { read open };
EOF

make -f /usr/share/selinux/devel/Makefile
semodule -i my_container.pp
```
