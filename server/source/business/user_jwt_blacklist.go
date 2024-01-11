package business

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/model/business"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"gorm.io/gorm"
)

const initOrderBusJwtBlacklist = initOrderMembers + 1

type initBusJwtBlacklist struct{}

// auto run
func init() {
	system.RegisterInit(initOrderBusJwtBlacklist, &initBusJwtBlacklist{})
}

func (i *initBusJwtBlacklist) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&business.BusJwtBlacklist{})
}

func (i *initBusJwtBlacklist) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&business.BusJwtBlacklist{})
}

func (i initBusJwtBlacklist) InitializerName() string {
	return business.BusJwtBlacklist{}.TableName()
}

func (i *initBusJwtBlacklist) InitializeData(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (i *initBusJwtBlacklist) DataInserted(ctx context.Context) bool {
	return true
}
