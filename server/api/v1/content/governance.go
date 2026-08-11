package content

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/api/v1/common"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	commonRequest "github.com/dyjh/order-food-mini-app/server/model/common/request"
	"github.com/dyjh/order-food-mini-app/server/model/common/response"
	contentRequest "github.com/dyjh/order-food-mini-app/server/model/content/request"
	contentResponse "github.com/dyjh/order-food-mini-app/server/model/content/response"
	"github.com/gin-gonic/gin"
)

// GovernanceApi 提供违规处理记录和异步任务接口处理能力。
type GovernanceApi struct{}

// ListGovernanceRecords 分页查询违规处理记录
// @Tags OrderFoodGovernance
// @Summary 分页查询违规处理记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query contentRequest.GovernanceRecordListQuery false "查询条件"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]contentResponse.GovernanceRecordSummary},msg=string} "获取成功"
// @Router /orderfood/governance-records [get]
func (api *GovernanceApi) ListGovernanceRecords(c *gin.Context) {
	var query contentRequest.GovernanceRecordListQuery
	if !apiCommon.BindAndVerify(c, &query, c.ShouldBindQuery) {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	var result response.Page[contentResponse.GovernanceRecordSummary]
	result, err := governanceService.ListRecords(c.Request.Context(), query, actor)
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
// @Success 200 {object} response.Response{data=contentResponse.GovernanceRecordDetail,msg=string} "获取成功"
// @Router /orderfood/governance-records/{recordId} [get]
func (api *GovernanceApi) GetGovernanceRecord(c *gin.Context) {
	path, ok := bindGovernanceRecordIDPath(c)
	if !ok {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, err := governanceService.GetRecord(c.Request.Context(), path.RecordID, actor)
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
// @Success 200 {object} response.Response{data=contentResponse.GovernanceJob,msg=string} "获取成功"
// @Router /orderfood/governance-jobs/{jobId} [get]
func (api *GovernanceApi) GetGovernanceJob(c *gin.Context) {
	path, ok := bindGovernanceJobIDPath(c)
	if !ok {
		return
	}
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, err := governanceService.GetJob(c.Request.Context(), path.JobID, actor)
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
// @Param data body contentRequest.GovernanceJobRetryInput true "失败项重试参数"
// @Success 200 {object} response.Response{data=contentResponse.GovernanceJob,msg=string} "操作成功"
// @Router /orderfood/governance-jobs/{jobId}/retry [post]
func (api *GovernanceApi) RetryGovernanceJob(c *gin.Context) {
	var header commonRequest.IdempotencyHeader
	if !apiCommon.BindAndVerify(c, &header, c.ShouldBindHeader) {
		return
	}
	path, ok := bindGovernanceJobIDPath(c)
	if !ok {
		return
	}
	var body contentRequest.GovernanceJobRetryInput
	if !apiCommon.BindAndVerify(c, &body, c.ShouldBindJSON) {
		return
	}
	header.Key = strings.TrimSpace(header.Key)
	body.Reason = strings.TrimSpace(body.Reason)
	actor, ok := apiCommon.CurrentAdminActor(c)
	if !ok {
		return
	}
	result, replayed, err := governanceService.RetryJob(
		c.Request.Context(), path.JobID, body, actor, header.Key,
	)
	if err != nil {
		response.FailWithBusinessError(err, c)
		return
	}
	apiCommon.SetReplayHeader(c, replayed)
	response.OkWithData(result, c)
}

// bindGovernanceRecordIDPath 绑定并规范化违规处理记录ID路径参数。
func bindGovernanceRecordIDPath(
	c *gin.Context,
) (contentRequest.GovernanceRecordIDPath, bool) {
	var path contentRequest.GovernanceRecordIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return contentRequest.GovernanceRecordIDPath{}, false
	}
	path.RecordID = strings.TrimSpace(path.RecordID)
	if path.RecordID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return contentRequest.GovernanceRecordIDPath{}, false
	}
	return path, true
}

// bindGovernanceJobIDPath 绑定并规范化违规处理任务ID路径参数。
func bindGovernanceJobIDPath(c *gin.Context) (contentRequest.GovernanceJobIDPath, bool) {
	var path contentRequest.GovernanceJobIDPath
	if !apiCommon.BindAndVerify(c, &path, c.ShouldBindUri) {
		return contentRequest.GovernanceJobIDPath{}, false
	}
	path.JobID = strings.TrimSpace(path.JobID)
	if path.JobID == "" {
		response.FailWithBusinessError(appErrors.AdminBadRequest.DefaultMsg(), c)
		return contentRequest.GovernanceJobIDPath{}, false
	}
	return path, true
}
