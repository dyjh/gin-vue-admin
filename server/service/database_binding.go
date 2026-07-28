package service

import (
	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"gorm.io/gorm"
)

// BindDatabase 将包级业务服务单例绑定到当前运行期数据库。
// 数据结构迁移与默认数据写入不属于此方法职责。
func (group *ServiceGroup) BindDatabase(db *gorm.DB) error {
	if group == nil || db == nil ||
		group.Points == nil ||
		group.Catalog == nil ||
		group.Recommendation == nil ||
		group.Content == nil ||
		group.AI == nil ||
		group.OfficialDish == nil ||
		group.Media == nil ||
		group.Governance == nil ||
		group.MealAdmin == nil ||
		group.Dashboard == nil ||
		group.AIUsage == nil ||
		group.AIUsage.Audit == nil ||
		group.Moderation == nil ||
		group.SuggestionCatalog == nil ||
		group.Subscription == nil ||
		group.WeChatConfig == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	if err := group.Content.BindDatabase(db); err != nil {
		return err
	}
	group.AI.DB = db
	group.Points.DB = db
	group.Catalog.DB = db
	group.Recommendation.DB = db
	group.OfficialDish.DB = db
	group.Media.DB = db
	group.Governance.DB = db
	group.MealAdmin.DB = db
	group.Dashboard.DB = db
	group.AIUsage.DB = db
	group.AIUsage.Audit.DB = db
	group.Moderation.DB = db
	group.SuggestionCatalog.DB = db
	group.Subscription.DB = db
	group.WeChatConfig.DB = db
	return nil
}
