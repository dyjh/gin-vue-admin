package meal

import (
	"errors"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"gorm.io/gorm"
)

// mealAdminLookupError 统一转换饭局域管理资源查询错误。
func mealAdminLookupError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.AdminNotFound.DefaultMsg()
	}
	return appErrors.AdminInternal.Wrap(err, "query admin meal resource")
}
