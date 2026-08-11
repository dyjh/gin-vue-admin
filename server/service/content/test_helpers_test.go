package content

import (
	"testing"

	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

func newOrderFoodTestDB(t *testing.T) *gorm.DB {
	return testutil.NewOrderFoodServiceDB(t, serviceCommon.DefaultRolePermissionMatrix())
}

func grantTestButtonPermission(t *testing.T, db *gorm.DB, authorityID uint, permission string) {
	testutil.GrantOrderFoodPermission(t, db, authorityID, permission)
}

func seedTestPermissionTemplates(t *testing.T, db *gorm.DB) {
	testutil.SeedOrderFoodPermissions(t, db, serviceCommon.DefaultRolePermissionMatrix())
}

func stringPointer(value string) *string {
	return &value
}
