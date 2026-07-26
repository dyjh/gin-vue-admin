package api

import (
	"io"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	response "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	"github.com/gin-gonic/gin"
)

// ContentApi 提供内容管理接口处理能力。
type ContentApi struct{}

// Metadata 获取菜品分类、标签、单位和运行时配置
// @Tags 内容管理
// @Summary 获取菜品分类、标签、单位和运行时配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Envelope{data=response.Metadata,msg=string} "获取成功"
// @Router /miniapp/v1/metadata [get]
func (*ContentApi) Metadata(c *gin.Context) {
	if _, ok := currentUser(c); !ok {
		return
	}
	runtime, err := runtimeService.Current(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	result, err := contentService.Metadata(c.Request.Context(), runtime)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// UploadImage 上传并同步审核用户图片
// @Tags 内容管理
// @Summary 上传并同步审核用户图片
// @Security ApiKeyAuth
// @accept multipart/form-data
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param scene formData string true "图片使用场景"
// @Param file formData file true "图片文件，最大 10MB"
// @Success 200 {object} response.Envelope{data=response.ImageUploadResult,msg=string} "操作成功"
// @Router /miniapp/v1/uploads/images [post]
func (*ContentApi) UploadImage(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	scene := c.PostForm("scene")
	file, err := c.FormFile("file")
	if err != nil || file.Size <= 0 || file.Size > 10<<20 {
		fail(c, appErrors.FrontInvalidImage.DefaultMsg())
		return
	}
	opened, err := file.Open()
	if err != nil {
		fail(c, appErrors.FrontInvalidImage.DefaultMsg())
		return
	}
	defer opened.Close()
	payload, err := io.ReadAll(io.LimitReader(opened, (10<<20)+1))
	if err != nil || len(payload) > 10<<20 {
		fail(c, appErrors.FrontInvalidImage.DefaultMsg())
		return
	}
	result, err := uploadService.UploadImage(
		c.Request.Context(), user.ID, scene, file.Header.Get("Content-Type"),
		payload, middleware.GetRequestID(c),
	)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// ListDishes 分页查询菜品列表
// @Tags 内容管理
// @Summary 分页查询菜品列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query frontRequest.DishListQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.Page[response.DishSummary],msg=string} "获取成功"
// @Router /miniapp/v1/dishes [get]
func (*ContentApi) ListDishes(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var query frontRequest.DishListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := contentService.ListDishes(c.Request.Context(), user.ID, query)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Dish 获取个人菜品详情
// @Tags 内容管理
// @Summary 获取个人菜品详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param dishId path string true "菜品ID"
// @Success 200 {object} response.Envelope{data=response.Dish,msg=string} "获取成功"
// @Router /miniapp/v1/dishes/{dishId} [get]
func (*ContentApi) Dish(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := contentService.Dish(c.Request.Context(), user.ID, c.Param("dishId"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// CreateDish 创建菜品
// @Tags 内容管理
// @Summary 创建菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.DishUpsertInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.Dish,msg=string} "操作成功"
// @Router /miniapp/v1/dishes [post]
func (*ContentApi) CreateDish(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.DishUpsertInput
	if !bindJSON(c, &input) {
		return
	}
	result, err := contentService.UpsertDish(c.Request.Context(), user.ID, nil, input)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// UpdateDish 更新菜品
// @Tags 内容管理
// @Summary 更新菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param dishId path string true "菜品ID"
// @Param data body frontRequest.DishUpsertInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.Dish,msg=string} "操作成功"
// @Router /miniapp/v1/dishes/{dishId} [put]
func (*ContentApi) UpdateDish(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.DishUpsertInput
	if !bindJSON(c, &input) {
		return
	}
	id := c.Param("dishId")
	result, err := contentService.UpsertDish(c.Request.Context(), user.ID, &id, input)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// DeleteDish 删除菜品
// @Tags 内容管理
// @Summary 删除菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param dishId path string true "菜品ID"
// @Success 200 {object} response.Envelope{data=response.DishDeleteResult,msg=string} "操作成功"
// @Router /miniapp/v1/dishes/{dishId} [delete]
func (*ContentApi) DeleteDish(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	recipes, meals, err := contentService.DeleteDish(c.Request.Context(), user.ID, c.Param("dishId"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.DishDeleteResult{
		Deleted:             true,
		AffectedRecipeCount: recipes,
		AffectedMealCount:   meals,
	}, c)
}

// Discoverability 设置菜品是否允许被发现
// @Tags 内容管理
// @Summary 设置菜品是否允许被发现
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param dishId path string true "菜品ID"
// @Param data body frontRequest.DiscoverabilityInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.Dish,msg=string} "操作成功"
// @Router /miniapp/v1/dishes/{dishId}/discoverability [put]
func (*ContentApi) Discoverability(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.DiscoverabilityInput
	if !bindJSON(c, &input) {
		return
	}
	result, err := contentService.SetDiscoverable(
		c.Request.Context(), user.ID, c.Param("dishId"), *input.Discoverable,
	)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Recommendations 分页查询平台推荐菜
// @Tags 内容管理
// @Summary 分页查询平台推荐菜
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query frontRequest.RecommendationListQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.Page[response.RecommendationSummary],msg=string} "获取成功"
// @Router /miniapp/v1/recommendations [get]
func (*ContentApi) Recommendations(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var query frontRequest.RecommendationListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := contentService.Recommendations(c.Request.Context(), user.ID, query)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// Recommendation 获取平台推荐菜详情
// @Tags 内容管理
// @Summary 获取平台推荐菜详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param recommendationId path string true "推荐ID"
// @Success 200 {object} response.Envelope{data=response.Recommendation,msg=string} "获取成功"
// @Router /miniapp/v1/recommendations/{recommendationId} [get]
func (*ContentApi) Recommendation(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := contentService.Recommendation(c.Request.Context(), user.ID, c.Param("recommendationId"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// CopyRecommendation 将平台推荐菜复制到个人菜品库
// @Tags 内容管理
// @Summary 将平台推荐菜复制到个人菜品库
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param recommendationId path string true "推荐ID"
// @Success 200 {object} response.Envelope{data=response.RecommendationCopyResult,msg=string} "操作成功"
// @Router /miniapp/v1/recommendations/{recommendationId}/copy [post]
func (*ContentApi) CopyRecommendation(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	dish, already, err := contentService.CopyRecommendation(c.Request.Context(), user.ID, c.Param("recommendationId"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.RecommendationCopyResult{
		Dish: dish, AlreadyCopied: already,
	}, c)
}

// Recipes 查询个人菜谱列表
// @Tags 内容管理
// @Summary 查询个人菜谱列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Envelope{data=response.RecipeListResult,msg=string} "获取成功"
// @Router /miniapp/v1/recipes [get]
func (*ContentApi) Recipes(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	list, err := contentService.Recipes(c.Request.Context(), user.ID)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.RecipeListResult{List: list}, c)
}

// Recipe 获取个人菜谱详情
// @Tags 内容管理
// @Summary 获取个人菜谱详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param recipeId path string true "菜谱ID"
// @Success 200 {object} response.Envelope{data=response.Recipe,msg=string} "获取成功"
// @Router /miniapp/v1/recipes/{recipeId} [get]
func (*ContentApi) Recipe(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := contentService.Recipe(c.Request.Context(), user.ID, c.Param("recipeId"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// CreateRecipe 创建菜谱
// @Tags 内容管理
// @Summary 创建菜谱
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.RecipeCreateInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.Recipe,msg=string} "操作成功"
// @Router /miniapp/v1/recipes [post]
func (*ContentApi) CreateRecipe(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.RecipeCreateInput
	if !bindJSON(c, &input) {
		return
	}
	result, err := contentService.CreateRecipe(c.Request.Context(), user.ID, input)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// UpdateRecipe 更新菜谱
// @Tags 内容管理
// @Summary 更新菜谱
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param recipeId path string true "菜谱ID"
// @Param data body frontRequest.RecipeUpdateInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.Recipe,msg=string} "操作成功"
// @Router /miniapp/v1/recipes/{recipeId} [put]
func (*ContentApi) UpdateRecipe(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.RecipeUpdateInput
	if !bindJSON(c, &input) {
		return
	}
	result, err := contentService.UpdateRecipe(c.Request.Context(), user.ID, c.Param("recipeId"), input)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// DeleteRecipe 删除菜谱
// @Tags 内容管理
// @Summary 删除菜谱
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param recipeId path string true "菜谱ID"
// @Success 200 {object} response.Envelope{data=response.DeletedResult,msg=string} "操作成功"
// @Router /miniapp/v1/recipes/{recipeId} [delete]
func (*ContentApi) DeleteRecipe(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	if err := contentService.DeleteRecipe(c.Request.Context(), user.ID, c.Param("recipeId")); err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(response.DeletedResult{Deleted: true}, c)
}

// AddRecipeDishes 批量添加菜品到菜谱
// @Tags 内容管理
// @Summary 批量添加菜品到菜谱
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param recipeId path string true "菜谱ID"
// @Param data body frontRequest.RecipeDishesInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.Recipe,msg=string} "操作成功"
// @Router /miniapp/v1/recipes/{recipeId}/dishes [post]
func (*ContentApi) AddRecipeDishes(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var input frontRequest.RecipeDishesInput
	if !bindJSON(c, &input) {
		return
	}
	result, err := contentService.AddRecipeDishes(c.Request.Context(), user.ID, c.Param("recipeId"), input.DishIDs)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// RemoveRecipeDish 从菜谱中移除指定菜品
// @Tags 内容管理
// @Summary 从菜谱中移除指定菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param recipeId path string true "菜谱ID"
// @Param dishId path string true "菜品ID"
// @Success 200 {object} response.Envelope{data=response.Recipe,msg=string} "操作成功"
// @Router /miniapp/v1/recipes/{recipeId}/dishes/{dishId} [delete]
func (*ContentApi) RemoveRecipeDish(c *gin.Context) {
	if _, ok := requireIdempotencyKey(c); !ok {
		return
	}
	user, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := contentService.RemoveRecipeDish(
		c.Request.Context(), user.ID, c.Param("recipeId"), c.Param("dishId"),
	)
	if err != nil {
		fail(c, err)
		return
	}
	response.OkWithData(result, c)
}
