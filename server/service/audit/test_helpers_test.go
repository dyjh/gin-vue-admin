package audit

import (
	"testing"

	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

func newOrderFoodTestDB(t *testing.T) *gorm.DB {
	return testutil.NewOrderFoodServiceDB(t, serviceCommon.DefaultRolePermissionMatrix())
}
