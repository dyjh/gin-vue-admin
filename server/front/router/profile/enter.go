package profile

import api "github.com/dyjh/order-food-mini-app/server/front/api"

type RouterGroup struct {
	ProfileRouter
}

var profileApi = api.ApiGroupApp.ProfileApiGroup.ProfileApi
