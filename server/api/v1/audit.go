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

// AuditApi 提供管理操作审计接口处理能力。
type AuditApi struct{}

// List 分页查询来干饭管理操作日志
// @Tags OrderFoodAudit
// @Summary 分页查询来干饭管理操作日志
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.AuditLogListQuery true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.AuditLog},msg=string} "获取成功"
// @Router /orderfood/audit-logs [get]
func (api *AuditApi) List(c *gin.Context) {
	var query orderfoodRequest.AuditLogListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	if !requirePermission(c, orderfoodService.PermissionAuditRead) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.AuditLog]
	result, err := orderFoodServiceGroup.Audit.List(c.Request.Context(), query)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// Detail 获取来干饭管理操作日志详情
// @Tags OrderFoodAudit
// @Summary 获取来干饭管理操作日志详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param auditLogId path string true "操作日志 ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.AuditLogDetail,msg=string} "获取成功"
// @Router /orderfood/audit-logs/{auditLogId} [get]
func (api *AuditApi) Detail(c *gin.Context) {
	var path orderfoodRequest.AuditLogIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return
	}
	path.AuditLogID = strings.TrimSpace(path.AuditLogID)
	if path.AuditLogID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return
	}
	if !requirePermission(c, orderfoodService.PermissionAuditRead) {
		return
	}
	result, err := orderFoodServiceGroup.Audit.Detail(c.Request.Context(), path.AuditLogID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}
