package router

import (
	"github.com/dyjh/order-food-mini-app/server/front/router/assist"
	"github.com/dyjh/order-food-mini-app/server/front/router/auth"
	"github.com/dyjh/order-food-mini-app/server/front/router/content"
	"github.com/dyjh/order-food-mini-app/server/front/router/engagement"
	"github.com/dyjh/order-food-mini-app/server/front/router/meal"
	"github.com/dyjh/order-food-mini-app/server/front/router/profile"
	"github.com/dyjh/order-food-mini-app/server/front/router/system"
)

type RouterGroup struct {
	Auth       auth.RouterGroup
	System     system.RouterGroup
	Profile    profile.RouterGroup
	Content    content.RouterGroup
	Engagement engagement.RouterGroup
	Meal       meal.RouterGroup
	Assist     assist.RouterGroup
}

var RouterGroupApp = new(RouterGroup)
