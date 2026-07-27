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

// GovernanceApi 提供违规处理记录和异步任务接口处理能力。
type GovernanceApi struct {
	service *orderfoodService.GovernanceService // 违规处理业务服务
}

// NewGovernanceApi 创建违规处理API实例。
func NewGovernanceApi(service *orderfoodService.GovernanceService) *GovernanceApi {
	return &GovernanceApi{service: service}
}

// ListGovernanceRecords 分页查询违规处理记录
// @Tags OrderFoodGovernance
// @Summary 分页查询违规处理记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query orderfoodRequest.GovernanceRecordListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]orderfoodResponse.GovernanceRecordSummary},msg=string} "获取成功"
// @Router /orderfood/governance-records [get]
func (api *GovernanceApi) ListGovernanceRecords(c *gin.Context) {
	var query orderfoodRequest.GovernanceRecordListQuery
	if !bindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	var result orderfoodResponse.Page[orderfoodResponse.GovernanceRecordSummary]
	result, err := api.service.ListRecords(c.Request.Context(), query, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetGovernanceRecord 获取违规处理记录详情
// @Tags OrderFoodGovernance
// @Summary 获取违规处理记录详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param recordId path string true "违规处理记录ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.GovernanceRecordDetail,msg=string} "获取成功"
// @Router /orderfood/governance-records/{recordId} [get]
func (api *GovernanceApi) GetGovernanceRecord(c *gin.Context) {
	path, ok := bindGovernanceRecordIDPath(c)
	if !ok {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	result, err := api.service.GetRecord(c.Request.Context(), path.RecordID, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// GetGovernanceJob 获取违规处理任务进度
// @Tags OrderFoodGovernance
// @Summary 获取违规处理任务进度
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param jobId path string true "违规处理任务ID"
// @Success 200 {object} response.Response{data=orderfoodResponse.GovernanceJob,msg=string} "获取成功"
// @Router /orderfood/governance-jobs/{jobId} [get]
func (api *GovernanceApi) GetGovernanceJob(c *gin.Context) {
	path, ok := bindGovernanceJobIDPath(c)
	if !ok {
		return
	}
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	result, err := api.service.GetJob(c.Request.Context(), path.JobID, actor)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	response.OkWithData(result, c)
}

// RetryGovernanceJob 重试违规处理任务失败项
// @Tags OrderFoodGovernance
// @Summary 重试违规处理任务失败项
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param jobId path string true "违规处理任务ID"
// @Param data body orderfoodRequest.GovernanceJobRetryInput true "失败项重试参数"
// @Success 200 {object} response.Response{data=orderfoodResponse.GovernanceJob,msg=string} "操作成功"
// @Router /orderfood/governance-jobs/{jobId}/retry [post]
func (api *GovernanceApi) RetryGovernanceJob(c *gin.Context) {
	var header orderfoodRequest.IdempotencyHeader
	if !bindAndVerify(c, &header, c.ShouldBindHeader) {
		return
	}
	path, ok := bindGovernanceJobIDPath(c)
	if !ok {
		return
	}
	var body orderfoodRequest.GovernanceJobRetryInput
	if !bindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	header.Key = strings.TrimSpace(header.Key)
	body.Reason = strings.TrimSpace(body.Reason)
	actor, ok := contentAdminActor(c)
	if !ok || !api.available(c) {
		return
	}
	result, replayed, err := api.service.RetryJob(
		c.Request.Context(), path.JobID, body, actor, header.Key,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	setAIReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// available 检查违规处理API是否已注入业务服务。
func (api *GovernanceApi) available(c *gin.Context) bool {
	if api != nil && api.service != nil {
		return true
	}
	response.FailWithBusinessError(appErrors.AdminInternal.DefaultMsg(), c)
	return false
}

// bindGovernanceRecordIDPath 绑定并规范化违规处理记录ID路径参数。
func bindGovernanceRecordIDPath(
	c *gin.Context,
) (orderfoodRequest.GovernanceRecordIDPath, bool) {
	var path orderfoodRequest.GovernanceRecordIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.GovernanceRecordIDPath{}, false
	}
	path.RecordID = strings.TrimSpace(path.RecordID)
	if path.RecordID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.GovernanceRecordIDPath{}, false
	}
	return path, true
}

// bindGovernanceJobIDPath 绑定并规范化违规处理任务ID路径参数。
func bindGovernanceJobIDPath(c *gin.Context) (orderfoodRequest.GovernanceJobIDPath, bool) {
	var path orderfoodRequest.GovernanceJobIDPath
	if !bindAndVerify(c, &path, c.ShouldBindUri) {
		return orderfoodRequest.GovernanceJobIDPath{}, false
	}
	path.JobID = strings.TrimSpace(path.JobID)
	if path.JobID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return orderfoodRequest.GovernanceJobIDPath{}, false
	}
	return path, true
}
