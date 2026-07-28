package service

import (
	"testing"

	"gorm.io/gorm"
)

func TestServiceGroupBindDatabaseBindsRuntimeServices(t *testing.T) {
	db := &gorm.DB{}
	group := NewServiceGroup(nil)

	if err := group.BindDatabase(db); err != nil {
		t.Fatal(err)
	}

	if group.Content.db != db ||
		group.AI.DB != db ||
		group.Points.DB != db ||
		group.Subscription.DB != db ||
		group.WeChatConfig.DB != db {
		t.Fatal("runtime services were not bound to the supplied database")
	}
}
