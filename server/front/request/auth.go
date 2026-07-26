package request

// WxLoginInput 表示微信登录输入参数。
type WxLoginInput struct {
	Code string `json:"code" binding:"required,min=1,max=256" checksql:"false"` // 编码
}

// ProfileUpdateInput 表示画像更新输入参数。
type ProfileUpdateInput struct {
	Nickname     string  `json:"nickname" binding:"required,max=30" checksql:"false"` // 昵称
	AvatarFileID *string `json:"avatarFileId" binding:"omitempty,max=64"`             // 头像文件ID
}
