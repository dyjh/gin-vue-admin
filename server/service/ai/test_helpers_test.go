package ai

import (
	"testing"

	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

func newOrderFoodTestDB(t *testing.T) *gorm.DB {
	return testutil.NewOrderFoodServiceDB(t, serviceCommon.DefaultRolePermissionMatrix())
}

func seedTestPermissionTemplates(t *testing.T, db *gorm.DB) {
	testutil.SeedOrderFoodPermissions(t, db, serviceCommon.DefaultRolePermissionMatrix())
}
