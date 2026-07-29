package v1

import (
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service"
	"github.com/gin-gonic/gin"
)

// UserApi 提供小程序用户接口处理能力。
type UserApi struct{}

// List 分页查询小程序用户
// @Tags OrderFoodUser
// @Summary 分页查询小程序用户
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.UserListQuery true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.UserSummary},msg=string} "获取成功"
// @Router /orderfood/users [get]
func (api *UserApi) List(c *gin.Context) {
	var query orderfoodRequest.UserListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	if !requirePermission(c, orderfoodService.PermissionUserRead) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.UserSummary]
	result, err := orderFoodServiceGroup.User.List(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// Detail 获取小程序用户详情
// @Tags OrderFoodUser
// @Summary 获取小程序用户详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param userId path string true "用户 ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.UserDetail,msg=string} "获取成功"
// @Router /orderfood/users/{userId} [get]
func (api *UserApi) Detail(c *gin.Context) {
	path, ok := bindUserIDPath(c)
	if !ok {
		return
	}
	if !requirePermission(c, orderfoodService.PermissionUserRead) {
		return
	}
	result, err := orderFoodServiceGroup.User.Detail(c.Request.Context(), path.UserID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// UpdateStatus 禁用或恢复小程序用户
// @Tags OrderFoodUser
// @Summary 禁用或恢复小程序用户
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param userId path string true "用户 ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.UpdateUserStatusBody true "状态变更参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.UserStatusResult,msg=string} "操作成功"
// @Router /orderfood/users/{userId}/status [put]
func (api *UserApi) UpdateStatus(c *gin.Context) {
	path, ok := bindUserIDPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.UpdateUserStatusBody
	if !bindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	if !requirePermission(c, orderfoodService.PermissionUserDisable) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := orderFoodServiceGroup.User.UpdateStatus(
		c.Request.Context(),
		actor,
		path.UserID,
		header.Key,
		body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	if replayed {
		c.Header("X-Idempotent-Replay", "true")
	}
	response.OkWithData(result, c)
}

// UpdateCapability 单独关闭或恢复小程序用户AI能力
// @Tags OrderFoodUser
// @Summary 单独关闭或恢复小程序用户AI能力
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param userId path string true "用户 ID"
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body orderfoodRequest.UpdateUserCapabilityBody true "能力状态变更参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.UserCapabilityResult,msg=string} "操作成功"
// @Router /orderfood/users/{userId}/capability [put]
func (api *UserApi) UpdateCapability(c *gin.Context) {
	path, ok := bindUserIDPath(c)
	if !ok {
		return
	}
	header, ok := bindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body orderfoodRequest.UpdateUserCapabilityBody
	if !bindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	if !requirePermission(c, orderfoodService.PermissionUserCapabilityUpdate) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := orderFoodServiceGroup.User.UpdateCapability(
		c.Request.Context(), actor, path.UserID, header.Key, body,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	if replayed {
		c.Header("X-Idempotent-Replay", "true")
	}
	response.OkWithData(result, c)
}

// PreferenceProfile 获取用户偏好画像
// @Tags OrderFoodUser
// @Summary 获取用户偏好画像
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param userId path string true "用户 ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.UserPreferenceProfile,msg=string} "获取成功"
// @Router /orderfood/users/{userId}/preference-profile [get]
func (api *UserApi) PreferenceProfile(c *gin.Context) {
	path, ok := bindUserIDPath(c)
	if !ok {
		return
	}
	if !requirePermission(c, orderfoodService.PermissionUserPreferenceRead) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok {
		return
	}
	result, err := orderFoodServiceGroup.User.PreferenceProfile(
		c.Request.Context(),
		path.UserID,
		orderfoodService.SensitiveAccess{
			AdministratorID:       actor.AdministratorID,
			AdministratorUsername: actor.Username,
			AdministratorNickname: actor.Nickname,
			AuthorityID:           actor.AuthorityID,
			Permission:            orderfoodService.PermissionUserPreferenceRead,
			RequestID:             actor.RequestID,
			SourceIPMasked:        actor.SourceIPMasked,
			UserAgentSummary:      actor.UserAgentSummary,
		},
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

func bindUserIDPath(c *gin.Context) (orderfoodRequest.UserIDPath, bool) {
	var path orderfoodRequest.UserIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.UserIDPath{}, false
	}
	if strings.TrimSpace(path.UserID) == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.UserIDPath{}, false
	}
	path.UserID = strings.TrimSpace(path.UserID)
	return path, true
}

func bindIdempotencyHeader(c *gin.Context) (orderfoodRequest.IdempotencyHeader, bool) {
	var header orderfoodRequest.IdempotencyHeader
	if !bindAndVerify(c, &header, c.ShouldBindHeader) {
		return orderfoodRequest.IdempotencyHeader{}, false
	}
	header.Key = strings.TrimSpace(header.Key)
	if header.Key == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.IdempotencyHeader{}, false
	}
	return header, true
}
