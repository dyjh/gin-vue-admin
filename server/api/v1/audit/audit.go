package audit

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	auditRequest "github.com/dyjh/order-food-mini-app/server/model/audit/request"
	auditResponse "github.com/dyjh/order-food-mini-app/server/model/audit/response"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/audit"
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
// @Param data query auditRequest.AuditLogListQuery true "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]auditResponse.AuditLog},msg=string} "获取成功"
// @Router /orderfood/audit-logs [get]
func (api *AuditApi) List(c *gin.Context) {
	var query auditRequest.AuditLogListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	if !apiCommon.RequirePermission(c, orderfoodService.PermissionAuditRead) {
		return
	}
	var result response.Page[auditResponse.AuditLog]
	result, err := auditService.List(c.Request.Context(), query)
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
// @Success 200 {object} response.Response{data=auditResponse.AuditLogDetail,msg=string} "获取成功"
// @Router /orderfood/audit-logs/{auditLogId} [get]
func (api *AuditApi) Detail(c *gin.Context) {
	var path auditRequest.AuditLogIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return
	}
	path.AuditLogID = strings.TrimSpace(path.AuditLogID)
	if path.AuditLogID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return
	}
	if !apiCommon.RequirePermission(c, orderfoodService.PermissionAuditRead) {
		return
	}
	result, err := auditService.Detail(c.Request.Context(), path.AuditLogID)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}
