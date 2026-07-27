package v1

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
)

// MediaApi 提供图片资源和图片审核记录接口处理能力。
type MediaApi struct {
	service *orderfoodService.MediaService // 图片资源业务服务
}

// NewMediaApi 创建图片资源API实例。
func NewMediaApi(service *orderfoodService.MediaService) *MediaApi {
	return &MediaApi{service: service}
}

// ListMedia 分页查询图片资源
// @Tags OrderFoodMedia
// @Summary 分页查询图片资源
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.MediaListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.MediaSummary},msg=string} "获取成功"
// @Router /orderfood/media [get]
func (api *MediaApi) ListMedia(c *gin.Context) {
	var query orderfoodRequest.MediaListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionMediaRead) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.MediaSummary]
	result, err := api.service.ListMedia(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetMedia 获取图片资源详情
// @Tags OrderFoodMedia
// @Summary 获取图片资源详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param fileId path string true "文件ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.MediaDetail,msg=string} "获取成功"
// @Router /orderfood/media/{fileId} [get]
func (api *MediaApi) GetMedia(c *gin.Context) {
	path, ok := bindMediaFileIDPath(c)
	if !ok || !requirePermission(c, orderfoodService.PermissionMediaRead) {
		return
	}
	result, err := api.service.GetMedia(c.Request.Context(), path.FileID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// ListModerationRecords 分页查询图片审核记录
// @Tags OrderFoodModerationRecord
// @Summary 分页查询图片审核记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.ModerationRecordListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.ModerationRecordSummary},msg=string} "获取成功"
// @Router /orderfood/moderation-records [get]
func (api *MediaApi) ListModerationRecords(c *gin.Context) {
	var query orderfoodRequest.ModerationRecordListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) ||
		!requirePermission(c, orderfoodService.PermissionModerationRead) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.ModerationRecordSummary]
	result, err := api.service.ListModerationRecords(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetModerationRecord 获取图片审核记录详情
// @Tags OrderFoodModerationRecord
// @Summary 获取图片审核记录详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param recordId path string true "审核记录ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.ModerationRecordDetail,msg=string} "获取成功"
// @Router /orderfood/moderation-records/{recordId} [get]
func (api *MediaApi) GetModerationRecord(c *gin.Context) {
	path, ok := bindModerationRecordIDPath(c)
	if !ok || !requirePermission(c, orderfoodService.PermissionModerationRead) {
		return
	}
	includeSensitive := orderFoodServiceGroup.Permission.Require(
		c.Request.Context(),
		utils.GetUserAuthorityId(c),
		orderfoodService.PermissionModerationSensitiveRead,
	) == nil
	result, err := api.service.GetModerationRecord(
		c.Request.Context(), path.RecordID, includeSensitive,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// bindMediaFileIDPath 绑定并规范化图片文件ID路径参数。
func bindMediaFileIDPath(c *gin.Context) (orderfoodRequest.MediaFileIDPath, bool) {
	var path orderfoodRequest.MediaFileIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.MediaFileIDPath{}, false
	}
	path.FileID = strings.TrimSpace(path.FileID)
	if path.FileID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.MediaFileIDPath{}, false
	}
	return path, true
}

// bindModerationRecordIDPath 绑定并规范化图片审核记录ID路径参数。
func bindModerationRecordIDPath(c *gin.Context) (orderfoodRequest.ModerationRecordIDPath, bool) {
	var path orderfoodRequest.ModerationRecordIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.ModerationRecordIDPath{}, false
	}
	path.RecordID = strings.TrimSpace(path.RecordID)
	if path.RecordID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.ModerationRecordIDPath{}, false
	}
	return path, true
}
