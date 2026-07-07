package initialize

import (
	"context"
	model "github.com/dyjh/order-food-mini-app/server/model/system"
	"github.com/dyjh/order-food-mini-app/server/plugin/plugin-tool/utils"
)

func Dictionary(ctx context.Context) {
	entities := []model.SysDictionary{}
	utils.RegisterDictionaries(entities...)
}
