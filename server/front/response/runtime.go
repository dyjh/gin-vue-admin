package response

// Profile 表示小程序用户资料响应数据。
type Profile struct {
	ID          string  `json:"id"`          // ID
	Nickname    string  `json:"nickname"`    // 昵称
	AvatarURL   *string `json:"avatarUrl"`   // 头像地址
	Points      *int64  `json:"points"`      // 积分
	CheckinDays int     `json:"checkinDays"` // 打卡天数
}

// ClientFeature 表示客户端功能响应数据。
type ClientFeature struct {
	Code          string  `json:"code"`          // 编码
	Title         string  `json:"title"`         // 标题
	ActionText    string  `json:"actionText"`    // 操作说明文本
	Description   string  `json:"description"`   // 说明
	Enabled       bool    `json:"enabled"`       // 是否启用
	PointCost     *int    `json:"pointCost"`     // 积分成本
	FreeQuota     *int    `json:"freeQuota"`     // 免费额度
	CostHint      *string `json:"costHint"`      // 积分消耗提示
	FreeQuotaHint *string `json:"freeQuotaHint"` // 免费额度提示
	SortOrder     int     `json:"sortOrder"`     // 排序值
}

// MealFinalResultSubscriptionConfig 表示饭局最终结果微信订阅入口配置。
type MealFinalResultSubscriptionConfig struct {
	Enabled    bool    `json:"enabled"`    // 是否展示并允许申请订阅
	TemplateID *string `json:"templateId"` // 微信订阅模板ID
	ButtonText string  `json:"buttonText"` // 客户端按钮文案
}

// RuntimeConfig 表示运行时配置响应数据。
type RuntimeConfig struct {
	PolicyVersion               int64                             `json:"policyVersion"`               // 策略版本
	EnhancedFeaturesEnabled     bool                              `json:"enhancedFeaturesEnabled"`     // 增强功能是否整体启用
	PointsEnabled               bool                              `json:"pointsEnabled"`               // 积分是否启用
	Features                    []ClientFeature                   `json:"features"`                    // 功能列表
	MealFinalResultSubscription MealFinalResultSubscriptionConfig `json:"mealFinalResultSubscription"` // 饭局最终结果订阅配置
}

// WxLoginResult 表示微信登录结果响应数据。
type WxLoginResult struct {
	AccessToken   string        `json:"accessToken"`   // 业务访问令牌
	ExpiresIn     int64         `json:"expiresIn"`     // 令牌有效秒数
	IsNewUser     bool          `json:"isNewUser"`     // 是否新用户
	Profile       Profile       `json:"profile"`       // 用户资料
	RuntimeConfig RuntimeConfig `json:"runtimeConfig"` // 运行时配置
}
