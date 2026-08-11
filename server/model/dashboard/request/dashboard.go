package request

// DashboardQuery 表示管理端运营概览查询条件。
type DashboardQuery struct {
	Range string `form:"range" json:"range" binding:"omitempty,oneof=today last7Days last30Days"` // 统计范围
}

// ApplyDefaults 补齐运营概览默认统计范围。
func (query *DashboardQuery) ApplyDefaults() {
	if query.Range == "" {
		query.Range = "today"
	}
}
