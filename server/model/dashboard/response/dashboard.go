package response

import (
	aiResponse "github.com/dyjh/order-food-mini-app/server/model/ai/response"
	"time"
)

// DashboardMetric 表示运营概览指标。
type DashboardMetric struct {
	Key         string                 `json:"key"`         // 指标编码
	Label       string                 `json:"label"`       // 指标名称
	Value       interface{}            `json:"value"`       // 指标值
	Unit        string                 `json:"unit"`        // 指标单位
	TargetPage  *string                `json:"targetPage"`  // 明细目标页面
	TargetQuery map[string]interface{} `json:"targetQuery"` // 明细页面查询条件
}

// DashboardAlert 表示运营概览异常提醒。
type DashboardAlert struct {
	Type        string                 `json:"type"`        // 异常类型
	Count       int64                  `json:"count"`       // 异常数量
	Level       string                 `json:"level"`       // 异常级别
	TargetPage  string                 `json:"targetPage"`  // 处理目标页面
	TargetQuery map[string]interface{} `json:"targetQuery"` // 处理页面查询条件
}

// DashboardData 表示运营概览响应数据。
type DashboardData struct {
	Range           string                                `json:"range"`           // 统计范围
	Timezone        string                                `json:"timezone"`        // 统计时区
	RangeStart      time.Time                             `json:"rangeStart"`      // 统计起始时间
	RangeEnd        time.Time                             `json:"rangeEnd"`        // 统计结束时间
	GeneratedAt     time.Time                             `json:"generatedAt"`     // 数据生成时间
	Currency        string                                `json:"currency"`        // 成本币种
	Metrics         []DashboardMetric                     `json:"metrics"`         // 运营指标
	CapabilityState aiResponse.PlatformCapabilitySnapshot `json:"capabilityState"` // 平台能力有效状态
	Alerts          []DashboardAlert                      `json:"alerts"`          // 异常提醒
}
