// Package policies 定义资源配置策略
package policies

import (
	"time"
)

// Policy 资源策略定义
type Policy struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Rules       []PolicyRule      `json:"rules"`
	Defaults    DefaultSettings   `json:"defaults"`
	Constraints ResourceConstraints `json:"constraints"`
}

// PolicyRule 策略规则
type PolicyRule struct {
	Name        string            `json:"name"`
	Condition   Condition         `json:"condition"`
	Action      Action            `json:"action"`
	Priority    int               `json:"priority"`
	Enabled     bool              `json:"enabled"`
}

// Condition 触发条件
type Condition struct {
	Type       string            `json:"type"`       // cpu, memory, io, custom
	Operator   string            `json:"operator"`   // >, <, ==, >=, <=
	Value      interface{}       `json:"value"`
	Duration   time.Duration     `json:"duration"`   // 持续时间
	Attributes map[string]string `json:"attributes"` // 额外属性
}

// Action 触发动作
type Action struct {
	Type       string            `json:"type"`       // scale, limit, notify, evict
	Parameters map[string]interface{} `json:"parameters"`
}

// DefaultSettings 默认设置
type DefaultSettings struct {
	CPURequest    string `json:"cpuRequest"`    // "100m"
	CPULimit      string `json:"cpuLimit"`      // "500m"
	MemoryRequest string `json:"memoryRequest"` // "128Mi"
	MemoryLimit   string `json:"memoryLimit"`   // "512Mi"
	EvictionThreshold float64 `json:"evictionThreshold"` // 0.9
}

// ResourceConstraints 资源约束
type ResourceConstraints struct {
	MinCPU        string `json:"minCPU"`        // 最小 CPU
	MaxCPU        string `json:"maxCPU"`        // 最大 CPU
	MinMemory     string `json:"minMemory"`     // 最小内存
	MaxMemory     string `json:"maxMemory"`     // 最大内存
	MinReplicas   int    `json:"minReplicas"`   // 最小副本数
	MaxReplicas   int    `json:"maxReplicas"`   // 最大副本数
}

// 预定义策略

// ProductionPolicy 生产环境策略
var ProductionPolicy = Policy{
	Name:        "production",
	Description: "生产环境资源配置策略",
	Defaults: DefaultSettings{
		CPURequest:    "250m",
		CPULimit:      "2000m",
		MemoryRequest: "256Mi",
		MemoryLimit:   "2Gi",
		EvictionThreshold: 0.85,
	},
	Constraints: ResourceConstraints{
		MinCPU:      "100m",
		MaxCPU:      "16000m",
		MinMemory:   "128Mi",
		MaxMemory:   "64Gi",
		MinReplicas: 2,
		MaxReplicas: 100,
	},
	Rules: []PolicyRule{
		{
			Name: "high-memory-usage",
			Condition: Condition{
				Type:     "memory",
				Operator: ">",
				Value:    0.8,
				Duration: 5 * time.Minute,
			},
			Action: Action{
				Type: "scale",
				Parameters: map[string]interface{}{
					"replicas": 2,
				},
			},
			Priority: 10,
			Enabled:  true,
		},
		{
			Name: "oom-risk",
			Condition: Condition{
				Type:     "memory",
				Operator: ">",
				Value:    0.9,
				Duration: 30 * time.Second,
			},
			Action: Action{
				Type: "notify",
				Parameters: map[string]interface{}{
					"severity": "critical",
					"message":  "Pod at high risk of OOM",
				},
			},
			Priority: 100,
			Enabled:  true,
		},
	},
}

// DevelopmentPolicy 开发环境策略
var DevelopmentPolicy = Policy{
	Name:        "development",
	Description: "开发环境资源配置策略",
	Defaults: DefaultSettings{
		CPURequest:    "100m",
		CPULimit:      "500m",
		MemoryRequest: "128Mi",
		MemoryLimit:   "512Mi",
		EvictionThreshold: 0.95,
	},
	Constraints: ResourceConstraints{
		MinCPU:      "10m",
		MaxCPU:      "4000m",
		MinMemory:   "16Mi",
		MaxMemory:   "8Gi",
		MinReplicas: 1,
		MaxReplicas: 10,
	},
	Rules: []PolicyRule{
		{
			Name: "idle-cleanup",
			Condition: Condition{
				Type:     "cpu",
				Operator: "<",
				Value:    0.05,
				Duration: 30 * time.Minute,
			},
			Action: Action{
				Type: "scale",
				Parameters: map[string]interface{}{
					"replicas": 0,
				},
			},
			Priority: 5,
			Enabled:  false,
		},
	},
}

// JavaPolicy Java 应用策略
var JavaPolicy = Policy{
	Name:        "java-application",
	Description: "Java 应用资源配置策略",
	Defaults: DefaultSettings{
		CPURequest:    "500m",
		CPULimit:      "4000m",
		MemoryRequest: "1Gi",
		MemoryLimit:   "8Gi",
		EvictionThreshold: 0.8,
	},
	Constraints: ResourceConstraints{
		MinCPU:      "200m",
		MaxCPU:      "16000m",
		MinMemory:   "512Mi",
		MaxMemory:   "32Gi",
		MinReplicas: 1,
		MaxReplicas: 50,
	},
	Rules: []PolicyRule{
		{
			Name: "heap-usage-high",
			Condition: Condition{
				Type:     "jvm.heap.used",
				Operator: ">",
				Value:    0.85,
				Duration: 2 * time.Minute,
			},
			Action: Action{
				Type: "notify",
				Parameters: map[string]interface{}{
					"severity": "warning",
					"message":  "JVM heap usage high, consider GC tuning",
				},
			},
			Priority: 20,
			Enabled:  true,
		},
		{
			Name: "gc-pause-long",
			Condition: Condition{
				Type:     "jvm.gc.pause",
				Operator: ">",
				Value:    "500ms",
				Duration: 1 * time.Minute,
			},
			Action: Action{
				Type: "notify",
				Parameters: map[string]interface{}{
					"severity": "warning",
					"message":  "Long GC pauses detected",
				},
			},
			Priority: 30,
			Enabled:  true,
		},
	},
}

// DatabasePolicy 数据库应用策略
var DatabasePolicy = Policy{
	Name:        "database",
	Description: "数据库应用资源配置策略",
	Defaults: DefaultSettings{
		CPURequest:    "1000m",
		CPULimit:      "8000m",
		MemoryRequest: "4Gi",
		MemoryLimit:   "32Gi",
		EvictionThreshold: 0.75,
	},
	Constraints: ResourceConstraints{
		MinCPU:      "500m",
		MaxCPU:      "32000m",
		MinMemory:   "2Gi",
		MaxMemory:   "128Gi",
		MinReplicas: 1,
		MaxReplicas: 10,
	},
	Rules: []PolicyRule{
		{
			Name: "connection-pool-exhaustion",
			Condition: Condition{
				Type:     "db.connections",
				Operator: ">",
				Value:    0.9,
				Duration: 30 * time.Second,
			},
			Action: Action{
				Type: "notify",
				Parameters: map[string]interface{}{
					"severity": "critical",
					"message":  "Database connection pool near exhaustion",
				},
			},
			Priority: 90,
			Enabled:  true,
		},
	},
}

// GetPolicy 根据名称获取策略
func GetPolicy(name string) *Policy {
	switch name {
	case "production":
		return &ProductionPolicy
	case "development":
		return &DevelopmentPolicy
	case "java":
		return &JavaPolicy
	case "database":
		return &DatabasePolicy
	default:
		return &ProductionPolicy
	}
}

// GetAllPolicies 获取所有预定义策略
func GetAllPolicies() []Policy {
	return []Policy{
		ProductionPolicy,
		DevelopmentPolicy,
		JavaPolicy,
		DatabasePolicy,
	}
}
