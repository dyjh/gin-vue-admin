package content

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	contentRequest "github.com/dyjh/order-food-mini-app/server/model/content/request"
	contentResponse "github.com/dyjh/order-food-mini-app/server/model/content/response"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/content"
	"github.com/gin-gonic/gin"
)

// MediaApi 提供图片资源和图片审核记录接口处理能力。
type MediaApi struct{}

// ListMedia 分页查询图片资源
// @Tags OrderFoodMedia
// @Summary 分页查询图片资源
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query contentRequest.MediaListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]contentResponse.MediaSummary},msg=string} "获取成功"
// @Router /orderfood/media [get]
func (api *MediaApi) ListMedia(c *gin.Context) {
	var query contentRequest.MediaListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, orderfoodService.PermissionMediaRead) {
		return
	}
	var result response.Page[contentResponse.MediaSummary]
	result, err := mediaService.ListMedia(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=contentResponse.MediaDetail,msg=string} "获取成功"
// @Router /orderfood/media/{fileId} [get]
func (api *MediaApi) GetMedia(c *gin.Context) {
	path, ok := bindMediaFileIDPath(c)
	if !ok || !apiCommon.RequirePermission(c, orderfoodService.PermissionMediaRead) {
		return
	}
	result, err := mediaService.GetMedia(c.Request.Context(), path.FileID)
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
// @Param data query contentRequest.ModerationRecordListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]contentResponse.ModerationRecordSummary},msg=string} "获取成功"
// @Router /orderfood/moderation-records [get]
func (api *MediaApi) ListModerationRecords(c *gin.Context) {
	var query contentRequest.ModerationRecordListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) ||
		!apiCommon.RequirePermission(c, serviceCommon.PermissionModerationRead) {
		return
	}
	var result response.Page[contentResponse.ModerationRecordSummary]
	result, err := mediaService.ListModerationRecords(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=contentResponse.ModerationRecordDetail,msg=string} "获取成功"
// @Router /orderfood/moderation-records/{recordId} [get]
func (api *MediaApi) GetModerationRecord(c *gin.Context) {
	path, ok := bindModerationRecordIDPath(c)
	if !ok || !apiCommon.RequirePermission(c, serviceCommon.PermissionModerationRead) {
		return
	}
	includeSensitive := apiCommon.HasPermission(
		c,
		orderfoodService.PermissionModerationSensitiveRead,
	)
	result, err := mediaService.GetModerationRecord(
		c.Request.Context(), path.RecordID, includeSensitive,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// bindMediaFileIDPath 绑定并规范化图片文件ID路径参数。
func bindMediaFileIDPath(c *gin.Context) (contentRequest.MediaFileIDPath, bool) {
	var path contentRequest.MediaFileIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return contentRequest.MediaFileIDPath{}, false
	}
	path.FileID = strings.TrimSpace(path.FileID)
	if path.FileID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return contentRequest.MediaFileIDPath{}, false
	}
	return path, true
}

// bindModerationRecordIDPath 绑定并规范化图片审核记录ID路径参数。
func bindModerationRecordIDPath(c *gin.Context) (contentRequest.ModerationRecordIDPath, bool) {
	var path contentRequest.ModerationRecordIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return contentRequest.ModerationRecordIDPath{}, false
	}
	path.RecordID = strings.TrimSpace(path.RecordID)
	if path.RecordID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return contentRequest.ModerationRecordIDPath{}, false
	}
	return path, true
}
