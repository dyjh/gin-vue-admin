package content

import "github.com/dyjh/order-food-mini-app/server/front/service"

// ApiGroup 聚合小程序内容接口。
type ApiGroup struct {
	ContentApi // 菜品、菜谱与上传接口
}

var (
	contentService = service.ServiceGroupApp.ExperienceServiceGroup.ContentService
	uploadService  = service.ServiceGroupApp.ExperienceServiceGroup.UploadService
	runtimeService = service.ServiceGroupApp.ExperienceServiceGroup.RuntimeService
)
