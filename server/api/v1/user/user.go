package user

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	userRequest "github.com/dyjh/order-food-mini-app/server/model/user/request"
	userResponse "github.com/dyjh/order-food-mini-app/server/model/user/response"
	serviceCommon "github.com/dyjh/order-food-mini-app/server/service/common"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/user"
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
// @Param data query userRequest.UserListQuery true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]userResponse.UserSummary},msg=string} "获取成功"
// @Router /orderfood/users [get]
func (api *UserApi) List(c *gin.Context) {
	var query userRequest.UserListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	if !apiCommon.RequirePermission(c, orderfoodService.PermissionUserRead) {
		return
	}
	var result response.Page[userResponse.UserSummary]
	result, err := userService.List(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=userResponse.UserDetail,msg=string} "获取成功"
// @Router /orderfood/users/{userId} [get]
func (api *UserApi) Detail(c *gin.Context) {
	path, ok := bindUserIDPath(c)
	if !ok {
		return
	}
	if !apiCommon.RequirePermission(c, orderfoodService.PermissionUserRead) {
		return
	}
	result, err := userService.Detail(c.Request.Context(), path.UserID)
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
// @Param data body userRequest.UpdateUserStatusBody true "状态变更参数"
// @Success 200 {object} response.Response{data=userResponse.UserStatusResult,msg=string} "操作成功"
// @Router /orderfood/users/{userId}/status [put]
func (api *UserApi) UpdateStatus(c *gin.Context) {
	path, ok := bindUserIDPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body userRequest.UpdateUserStatusBody
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	if !apiCommon.RequirePermission(c, orderfoodService.PermissionUserDisable) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := userService.UpdateStatus(
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
// @Param data body userRequest.UpdateUserCapabilityBody true "能力状态变更参数"
// @Success 200 {object} response.Response{data=userResponse.UserCapabilityResult,msg=string} "操作成功"
// @Router /orderfood/users/{userId}/capability [put]
func (api *UserApi) UpdateCapability(c *gin.Context) {
	path, ok := bindUserIDPath(c)
	if !ok {
		return
	}
	header, ok := apiCommon.BindIdempotencyHeader(c)
	if !ok {
		return
	}
	var body userRequest.UpdateUserCapabilityBody
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	if !apiCommon.RequirePermission(c, orderfoodService.PermissionUserCapabilityUpdate) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := userService.UpdateCapability(
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
// @Success 200 {object} response.Response{data=userResponse.UserPreferenceProfile,msg=string} "获取成功"
// @Router /orderfood/users/{userId}/preference-profile [get]
func (api *UserApi) PreferenceProfile(c *gin.Context) {
	path, ok := bindUserIDPath(c)
	if !ok {
		return
	}
	if !apiCommon.RequirePermission(c, orderfoodService.PermissionUserPreferenceRead) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, err := userService.PreferenceProfile(
		c.Request.Context(),
		path.UserID,
		serviceCommon.SensitiveAccess{
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

func bindUserIDPath(c *gin.Context) (userRequest.UserIDPath, bool) {
	var path userRequest.UserIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return userRequest.UserIDPath{}, false
	}
	if strings.TrimSpace(path.UserID) == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return userRequest.UserIDPath{}, false
	}
	path.UserID = strings.TrimSpace(path.UserID)
	return path, true
}
