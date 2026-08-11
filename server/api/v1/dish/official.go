package dish

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"io"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	dishRequest "github.com/dyjh/order-food-mini-app/server/model/dish/request"
	dishResponse "github.com/dyjh/order-food-mini-app/server/model/dish/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/dish"
	"github.com/gin-gonic/gin"
)

// OfficialDishApi 提供官方菜品和官方封面接口处理能力。
type OfficialDishApi struct{}

// ListOfficialDishes 分页查询官方菜品
// @Tags OrderFoodOfficialDish
// @Summary 分页查询官方菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query dishRequest.OfficialDishListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]dishResponse.OfficialDishSummary},msg=string} "获取成功"
// @Router /orderfood/official-dishes [get]
func (api *OfficialDishApi) ListOfficialDishes(c *gin.Context) {
	var query dishRequest.OfficialDishListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionOfficialDishRead) {
		return
	}
	var result response.Page[dishResponse.OfficialDishSummary]
	result, err := officialDishService.ListOfficialDishes(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetOfficialDish 获取官方菜品详情
// @Tags OrderFoodOfficialDish
// @Summary 获取官方菜品详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param dishId path string true "官方菜品ID"
// @Success 200 {object} response.Response{data=dishResponse.OfficialDishDetail,msg=string} "获取成功"
// @Router /orderfood/official-dishes/{dishId} [get]
func (api *OfficialDishApi) GetOfficialDish(c *gin.Context) {
	path, ok := bindOfficialDishIDPath(c)
	if !ok || !apiCommon.RequirePermission(c, orderfoodService.PermissionOfficialDishRead) {
		return
	}
	result, err := officialDishService.GetOfficialDish(c.Request.Context(), path.DishID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// CreateOfficialDish 创建官方菜品
// @Tags OrderFoodOfficialDish
// @Summary 创建官方菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body dishRequest.OfficialDishCreateInput true "官方菜品参数"
// @Success 200 {object} response.Response{data=dishResponse.OfficialDishDetail,msg=string} "操作成功"
// @Router /orderfood/official-dishes [post]
func (api *OfficialDishApi) CreateOfficialDish(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body dishRequest.OfficialDishCreateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionOfficialDishCreate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := officialDishService.CreateOfficialDish(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// UpdateOfficialDish 编辑官方菜品并自动下线在线推荐
// @Tags OrderFoodOfficialDish
// @Summary 编辑官方菜品并自动下线在线推荐
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param dishId path string true "官方菜品ID"
// @Param data body dishRequest.OfficialDishUpdateInput true "官方菜品编辑参数"
// @Success 200 {object} response.Response{data=dishResponse.OfficialDishMutationResult,msg=string} "操作成功"
// @Router /orderfood/official-dishes/{dishId} [put]
func (api *OfficialDishApi) UpdateOfficialDish(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindOfficialDishIDPath(c)
	if !ok {
		return
	}
	var body dishRequest.OfficialDishUpdateInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionOfficialDishUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := officialDishService.UpdateOfficialDish(
		c.Request.Context(), path.DishID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// DeleteOfficialDish 软删除官方菜品并自动下线在线推荐
// @Tags OrderFoodOfficialDish
// @Summary 软删除官方菜品并自动下线在线推荐
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param dishId path string true "官方菜品ID"
// @Param data body dishRequest.OfficialDishDeleteInput true "官方菜品删除参数"
// @Success 200 {object} response.Response{data=dishResponse.OfficialDishDeleteResult,msg=string} "操作成功"
// @Router /orderfood/official-dishes/{dishId} [delete]
func (api *OfficialDishApi) DeleteOfficialDish(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindOfficialDishIDPath(c)
	if !ok {
		return
	}
	var body dishRequest.OfficialDishDeleteInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionOfficialDishDelete) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := officialDishService.DeleteOfficialDish(
		c.Request.Context(), path.DishID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListOfficialDishCovers 分页查询官方菜品封面
// @Tags OrderFoodOfficialDish
// @Summary 分页查询官方菜品封面
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query dishRequest.OfficialDishCoverListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]dishResponse.OfficialDishCoverSummary},msg=string} "获取成功"
// @Router /orderfood/official-dish-covers [get]
func (api *OfficialDishApi) ListOfficialDishCovers(c *gin.Context) {
	var query dishRequest.OfficialDishCoverListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionOfficialDishRead) {
		return
	}
	result, err := officialDishService.ListOfficialDishCovers(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// UploadOfficialDishCover 上传不调用图片审核供应商的官方菜品封面
// @Tags OrderFoodOfficialDish
// @Summary 上传官方菜品封面
// @Security ApiKeyAuth
// @accept multipart/form-data
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param file formData file true "图片文件"
// @Success 200 {object} response.Response{data=dishResponse.OfficialDishCoverSummary,msg=string} "上传成功"
// @Router /orderfood/official-dish-covers [post]
func (api *OfficialDishApi) UploadOfficialDishCover(c *gin.Context) {
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok || !apiCommon.RequirePermission(c, orderfoodService.PermissionOfficialDishCoverUpload) {
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader == nil || fileHeader.Size <= 0 ||
		fileHeader.Size > orderfoodService.MaxOfficialCoverBytes() {
		response.FailWithBusinessError(appErrors.AdminInvalidImage.DefaultMsg(), c)
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.FailWithBusinessError(appErrors.AdminInvalidImage.DefaultMsg(), c)
		return
	}
	defer file.Close()
	payload, err := io.ReadAll(io.LimitReader(file, orderfoodService.MaxOfficialCoverBytes()+1))
	if err != nil || int64(len(payload)) != fileHeader.Size ||
		int64(len(payload)) > orderfoodService.MaxOfficialCoverBytes() {
		response.FailWithBusinessError(appErrors.AdminInvalidImage.DefaultMsg(), c)
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := officialDishService.UploadOfficialDishCover(
		c.Request.Context(),
		actor,
		header.Key,
		fileHeader.Filename,
		fileHeader.Header.Get("Content-Type"),
		payload,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// bindOfficialDishIDPath 绑定并规范化官方菜品ID路径参数。
func bindOfficialDishIDPath(c *gin.Context) (dishRequest.OfficialDishIDPath, bool) {
	var path dishRequest.OfficialDishIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return dishRequest.OfficialDishIDPath{}, false
	}
	path.DishID = strings.TrimSpace(path.DishID)
	if path.DishID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return dishRequest.OfficialDishIDPath{}, false
	}
	return path, true
}
