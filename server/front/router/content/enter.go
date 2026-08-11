package content

import api "github.com/dyjh/order-food-mini-app/server/front/api"

type RouterGroup struct {
	ContentRouter
}

var contentApi = api.ApiGroupApp.ContentApiGroup.ContentApi
