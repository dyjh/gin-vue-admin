package orderfood

import (
	"io"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/gin-gonic/gin"
)

// OfficialDishApi 提供官方菜品和官方封面接口处理能力。
type OfficialDishApi struct {
	service *orderfoodService.OfficialDishService // 官方菜品业务服务
}

// NewOfficialDishApi 创建官方菜品API实例。
func NewOfficialDishApi(service *orderfoodService.OfficialDishService) *OfficialDishApi {
	return &OfficialDishApi{service: service}
}

// ListOfficialDishes 分页查询官方菜品
// @Tags OrderFoodOfficialDish
// @Summary 分页查询官方菜品
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.OfficialDishListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.OfficialDishSummary},msg=string} "获取成功"
// @Router /orderfood/official-dishes [get]
func (api *OfficialDishApi) ListOfficialDishes(c *gin.Context) {
	var query orderfoodRequest.OfficialDishListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionOfficialDishRead) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.OfficialDishSummary]
	result, err := api.service.ListOfficialDishes(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=orderfoodResponse.OfficialDishDetail,msg=string} "获取成功"
// @Router /orderfood/official-dishes/{dishId} [get]
func (api *OfficialDishApi) GetOfficialDish(c *gin.Context) {
	path, ok := bindOfficialDishIDPath(c)
	if !ok || !requirePermission(c, orderfoodService.PermissionOfficialDishRead) {
		return
	}
	result, err := api.service.GetOfficialDish(c.Request.Context(), path.DishID)
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
// @Param data body orderfoodRequest.OfficialDishCreateInput true "官方菜品参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.OfficialDishDetail,msg=string} "操作成功"
// @Router /orderfood/official-dishes [post]
func (api *OfficialDishApi) CreateOfficialDish(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.OfficialDishCreateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionOfficialDishCreate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.CreateOfficialDish(
		c.Request.Context(), actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.OfficialDishUpdateInput true "官方菜品编辑参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.OfficialDishMutationResult,msg=string} "操作成功"
// @Router /orderfood/official-dishes/{dishId} [put]
func (api *OfficialDishApi) UpdateOfficialDish(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindOfficialDishIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.OfficialDishUpdateInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionOfficialDishUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UpdateOfficialDish(
		c.Request.Context(), path.DishID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
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
// @Param data body orderfoodRequest.OfficialDishDeleteInput true "官方菜品删除参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.OfficialDishDeleteResult,msg=string} "操作成功"
// @Router /orderfood/official-dishes/{dishId} [delete]
func (api *OfficialDishApi) DeleteOfficialDish(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	path, ok := bindOfficialDishIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.OfficialDishDeleteInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) ||
		!requirePermission(c, orderfoodService.PermissionOfficialDishDelete) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.DeleteOfficialDish(
		c.Request.Context(), path.DishID, actor, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// ListOfficialDishCovers 分页查询官方菜品封面
// @Tags OrderFoodOfficialDish
// @Summary 分页查询官方菜品封面
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.OfficialDishCoverListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.OfficialDishCoverSummary},msg=string} "获取成功"
// @Router /orderfood/official-dish-covers [get]
func (api *OfficialDishApi) ListOfficialDishCovers(c *gin.Context) {
	var query orderfoodRequest.OfficialDishCoverListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionOfficialDishRead) {
		return
	}
	result, err := api.service.ListOfficialDishCovers(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=orderfoodResponse.OfficialDishCoverSummary,msg=string} "上传成功"
// @Router /orderfood/official-dish-covers [post]
func (api *OfficialDishApi) UploadOfficialDishCover(c *gin.Context) {
	header, ok := bindIdempotencyHeader(c)
	if !ok || !requirePermission(c, orderfoodService.PermissionOfficialDishCoverUpload) {
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
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := api.service.UploadOfficialDishCover(
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
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// bindOfficialDishIDPath 绑定并规范化官方菜品ID路径参数。
func bindOfficialDishIDPath(c *gin.Context) (orderfoodRequest.OfficialDishIDPath, bool) {
	var path orderfoodRequest.OfficialDishIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.OfficialDishIDPath{}, false
	}
	path.DishID = strings.TrimSpace(path.DishID)
	if path.DishID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.OfficialDishIDPath{}, false
	}
	return path, true
}
