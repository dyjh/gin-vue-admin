package request

// WeChatConfigQuery 表示微信配置空查询条件。
type WeChatConfigQuery struct{}

// WeChatConfigUpdateInput 表示微信小程序配置直接保存参数。
type WeChatConfigUpdateInput struct {
	AppID           string  `json:"appId" binding:"required,len=18" checksql:"false"`         // 微信小程序AppID
	AppSecret       *string `json:"appSecret" binding:"omitempty,len=32" checksql:"false"`    // 新AppSecret，留空时保持原值
	Reason          string  `json:"reason" binding:"required,min=2,max=200" checksql:"false"` // 修改原因
	ExpectedVersion int64   `json:"expectedVersion" binding:"min=0"`                          // 当前配置版本，首次配置传0
}
