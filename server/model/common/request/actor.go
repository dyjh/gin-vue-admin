package request

// AdminActor 表示后台管理操作的管理员上下文。
type AdminActor struct {
	AdministratorID  uint    // 管理员ID
	AuthorityID      uint    // 角色ID
	Username         string  // 用户名
	Nickname         *string // 昵称
	RequestID        string  // 请求ID
	SourceIPMasked   string  // 脱敏后的来源IP
	UserAgentSummary string  // 用户代理摘要
}
