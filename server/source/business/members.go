package business

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/model/business"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"gorm.io/gorm"
)

const initOrderMembers = system.InitOrderInternal + 1

type initMembers struct{}

// auto run
func init() {
	system.RegisterInit(initOrderMembers, &initMembers{})
}

func (i *initMembers) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&business.Members{})
}

func (i *initMembers) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&business.Members{})
}

func (i initMembers) InitializerName() string {
	return business.Members{}.TableName()
}

func (i *initMembers) InitializeData(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (i *initMembers) DataInserted(ctx context.Context) bool {
	return true
}
