// Package monitoring 定义监控指标
package monitoring

import (
	"fmt"
	"strings"
)

// MetricType 指标类型
type MetricType string

const (
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeCounter   MetricType = "counter"
	MetricTypeHistogram MetricType = "histogram"
)

// Metric 指标定义
type Metric struct {
	Name        string      `json:"name"`
	Type        MetricType  `json:"type"`
	Help        string      `json:"help"`
	Labels      []string    `json:"labels"`
	Buckets     []float64   `json:"buckets,omitempty"` // for histogram
	Unit        string      `json:"unit"`
}

// CPU 指标
var CPUMetrics = []Metric{
	{
		Name:   "container_cpu_usage_seconds_total",
		Type:   MetricTypeCounter,
		Help:   "Total CPU usage in seconds",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "seconds",
	},
	{
		Name:   "container_cpu_cfs_periods_total",
		Type:   MetricTypeCounter,
		Help:   "Number of CFS scheduler periods",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "count",
	},
	{
		Name:   "container_cpu_cfs_throttled_periods_total",
		Type:   MetricTypeCounter,
		Help:   "Number of throttled CFS periods",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "count",
	},
	{
		Name:   "container_cpu_cfs_throttled_seconds_total",
		Type:   MetricTypeCounter,
		Help:   "Total time CPU was throttled",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "seconds",
	},
	{
		Name:   "node_cpu_pressure",
		Type:   MetricTypeGauge,
		Help:   "CPU pressure stall information",
		Labels: []string{"node", "type"}, // type: some, full
		Unit:   "ratio",
	},
}

// Memory 指标
var MemoryMetrics = []Metric{
	{
		Name:   "container_memory_usage_bytes",
		Type:   MetricTypeGauge,
		Help:   "Current memory usage in bytes",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "bytes",
	},
	{
		Name:   "container_memory_max_usage_bytes",
		Type:   MetricTypeGauge,
		Help:   "Maximum memory usage in bytes",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "bytes",
	},
	{
		Name:   "container_memory_cache",
		Type:   MetricTypeGauge,
		Help:   "Page cache memory in bytes",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "bytes",
	},
	{
		Name:   "container_memory_rss",
		Type:   MetricTypeGauge,
		Help:   "RSS memory in bytes",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "bytes",
	},
	{
		Name:   "container_memory_working_set_bytes",
		Type:   MetricTypeGauge,
		Help:   "Working set memory (usage - inactive file)",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "bytes",
	},
	{
		Name:   "container_memory_failcnt",
		Type:   MetricTypeCounter,
		Help:   "Memory limit hit count",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "count",
	},
	{
		Name:   "node_memory_pressure",
		Type:   MetricTypeGauge,
		Help:   "Memory pressure stall information",
		Labels: []string{"node", "type"},
		Unit:   "ratio",
	},
}

// I/O 指标
var IOMetrics = []Metric{
	{
		Name:   "container_fs_usage_bytes",
		Type:   MetricTypeGauge,
		Help:   "Filesystem usage in bytes",
		Labels: []string{"container", "pod", "namespace", "device"},
		Unit:   "bytes",
	},
	{
		Name:   "container_fs_limit_bytes",
		Type:   MetricTypeGauge,
		Help:   "Filesystem limit in bytes",
		Labels: []string{"container", "pod", "namespace", "device"},
		Unit:   "bytes",
	},
	{
		Name:   "container_fs_reads_total",
		Type:   MetricTypeCounter,
		Help:   "Total number of reads",
		Labels: []string{"container", "pod", "namespace", "device"},
		Unit:   "count",
	},
	{
		Name:   "container_fs_writes_total",
		Type:   MetricTypeCounter,
		Help:   "Total number of writes",
		Labels: []string{"container", "pod", "namespace", "device"},
		Unit:   "count",
	},
	{
		Name:   "container_fs_read_seconds_total",
		Type:   MetricTypeCounter,
		Help:   "Total time spent reading",
		Labels: []string{"container", "pod", "namespace", "device"},
		Unit:   "seconds",
	},
	{
		Name:   "container_fs_write_seconds_total",
		Type:   MetricTypeCounter,
		Help:   "Total time spent writing",
		Labels: []string{"container", "pod", "namespace", "device"},
		Unit:   "seconds",
	},
	{
		Name:   "node_io_pressure",
		Type:   MetricTypeGauge,
		Help:   "I/O pressure stall information",
		Labels: []string{"node", "type"},
		Unit:   "ratio",
	},
}

// Network 指标
var NetworkMetrics = []Metric{
	{
		Name:   "container_network_receive_bytes_total",
		Type:   MetricTypeCounter,
		Help:   "Total bytes received",
		Labels: []string{"container", "pod", "namespace", "interface"},
		Unit:   "bytes",
	},
	{
		Name:   "container_network_transmit_bytes_total",
		Type:   MetricTypeCounter,
		Help:   "Total bytes transmitted",
		Labels: []string{"container", "pod", "namespace", "interface"},
		Unit:   "bytes",
	},
	{
		Name:   "container_network_receive_packets_total",
		Type:   MetricTypeCounter,
		Help:   "Total packets received",
		Labels: []string{"container", "pod", "namespace", "interface"},
		Unit:   "count",
	},
	{
		Name:   "container_network_transmit_packets_total",
		Type:   MetricTypeCounter,
		Help:   "Total packets transmitted",
		Labels: []string{"container", "pod", "namespace", "interface"},
		Unit:   "count",
	},
	{
		Name:   "container_network_receive_errors_total",
		Type:   MetricTypeCounter,
		Help:   "Total receive errors",
		Labels: []string{"container", "pod", "namespace", "interface"},
		Unit:   "count",
	},
	{
		Name:   "container_network_transmit_errors_total",
		Type:   MetricTypeCounter,
		Help:   "Total transmit errors",
		Labels: []string{"container", "pod", "namespace", "interface"},
		Unit:   "count",
	},
}

// OOM 指标
var OOMMetrics = []Metric{
	{
		Name:   "container_oom_events_total",
		Type:   MetricTypeCounter,
		Help:   "Total OOM events",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "count",
	},
	{
		Name:   "container_oom_kills_total",
		Type:   MetricTypeCounter,
		Help:   "Total processes killed by OOM",
		Labels: []string{"container", "pod", "namespace"},
		Unit:   "count",
	},
}

// 聚合指标
var AggregateMetrics = []Metric{
	{
		Name:   "pod_resource_efficiency",
		Type:   MetricTypeGauge,
		Help:   "Resource utilization efficiency (usage/request)",
		Labels: []string{"pod", "namespace", "resource"},
		Unit:   "ratio",
	},
	{
		Name:   "pod_resource_saturation",
		Type:   MetricTypeGauge,
		Help:   "Resource saturation (usage/limit)",
		Labels: []string{"pod", "namespace", "resource"},
		Unit:   "ratio",
	},
	{
		Name:   "node_resource_fragmentation",
		Type:   MetricTypeGauge,
		Help:   "Node resource fragmentation score",
		Labels: []string{"node", "resource"},
		Unit:   "ratio",
	},
	{
		Name:   "cluster_resource_utilization",
		Type:   MetricTypeGauge,
		Help:   "Cluster-wide resource utilization",
		Labels: []string{"resource"},
		Unit:   "ratio",
	},
}

// GetAllMetrics 获取所有指标定义
func GetAllMetrics() []Metric {
	var all []Metric
	all = append(all, CPUMetrics...)
	all = append(all, MemoryMetrics...)
	all = append(all, IOMetrics...)
	all = append(all, NetworkMetrics...)
	all = append(all, OOMMetrics...)
	all = append(all, AggregateMetrics...)
	return all
}

// GetMetricsByType 按类型获取指标
func GetMetricsByType(metricType string) []Metric {
	switch strings.ToLower(metricType) {
	case "cpu":
		return CPUMetrics
	case "memory":
		return MemoryMetrics
	case "io", "disk":
		return IOMetrics
	case "network":
		return NetworkMetrics
	case "oom":
		return OOMMetrics
	case "aggregate":
		return AggregateMetrics
	default:
		return nil
	}
}

// ToPrometheusFormat 转换为 Prometheus 格式
func (m *Metric) ToPrometheusFormat() string {
	var sb strings.Builder

	// HELP
	sb.WriteString(fmt.Sprintf("# HELP %s %s\n", m.Name, m.Help))

	// TYPE
	sb.WriteString(fmt.Sprintf("# TYPE %s %s\n", m.Name, m.Type))

	return sb.String()
}

// MetricValue 指标值
type MetricValue struct {
	Metric    string            `json:"metric"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp int64             `json:"timestamp"`
}

// MetricSeries 指标时间序列
type MetricSeries struct {
	Metric   Metric         `json:"metric"`
	Values   []MetricValue  `json:"values"`
	Labels   map[string]string `json:"labels"`
}

// CalculateEfficiency 计算资源效率
func CalculateEfficiency(usage, request float64) float64 {
	if request <= 0 {
		return 0
	}
	return usage / request
}

// CalculateSaturation 计算资源饱和度
func CalculateSaturation(usage, limit float64) float64 {
	if limit <= 0 {
		return 0
	}
	return usage / limit
}
