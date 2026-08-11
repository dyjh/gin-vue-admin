package engagement

import (
	apiCommon "github.com/dyjh/order-food-mini-app/server/front/api/common"
	frontRequest "github.com/dyjh/order-food-mini-app/server/front/request"
	response "github.com/dyjh/order-food-mini-app/server/front/response"
	"github.com/gin-gonic/gin"
)

// EngagementApi 提供打卡、积分与通知接口处理能力。
type EngagementApi struct{}

// CheckinCalendar 获取指定月份的做菜打卡日历
// @Tags 打卡、积分与通知
// @Summary 获取指定月份的做菜打卡日历
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query frontRequest.CheckinCalendarQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.CheckinCalendarResult,msg=string} "获取成功"
// @Router /miniapp/v1/checkins/calendar [get]
func (*EngagementApi) CheckinCalendar(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var query frontRequest.CheckinCalendarQuery
	if !apiCommon.BindQuery(c, &query) {
		return
	}
	dates, days, err := engagementService.CheckinCalendar(c.Request.Context(), user.ID, query.Month)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.CheckinCalendarResult{
		Month: query.Month, CheckedDates: dates, CheckinDays: days,
	}, c)
}

// Checkins 分页查询做菜打卡记录
// @Tags 打卡、积分与通知
// @Summary 分页查询做菜打卡记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query frontRequest.PageQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.Page[response.Checkin],msg=string} "获取成功"
// @Router /miniapp/v1/checkins [get]
func (*EngagementApi) Checkins(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var query frontRequest.PageQuery
	if !apiCommon.BindQuery(c, &query) {
		return
	}
	result, err := engagementService.Checkins(c.Request.Context(), user.ID, query)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// CreateCheckin 创建做菜打卡并结算积分奖励
// @Tags 打卡、积分与通知
// @Summary 创建做菜打卡并结算积分奖励
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param data body frontRequest.CheckinInput true "请求参数"
// @Success 200 {object} response.Envelope{data=response.CheckinCreateResult,msg=string} "操作成功"
// @Router /miniapp/v1/checkins [post]
func (*EngagementApi) CreateCheckin(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var input frontRequest.CheckinInput
	if !apiCommon.BindJSON(c, &input) {
		return
	}
	checkin, rewarded, amount, queued, err := engagementService.CreateCheckin(c.Request.Context(), user.ID, input)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.CheckinCreateResult{
		Checkin: checkin, RewardGranted: rewarded, RewardAmount: amount,
		PreferenceUpdateQueued: queued,
	}, c)
}

// PointsSummary 获取当前积分及本月收支汇总
// @Tags 打卡、积分与通知
// @Summary 获取当前积分及本月收支汇总
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Envelope{data=response.PointsSummaryResult,msg=string} "获取成功"
// @Router /miniapp/v1/points/summary [get]
func (*EngagementApi) PointsSummary(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	balance, earned, spent, err := engagementService.PointsSummary(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.PointsSummaryResult{
		Balance: balance, MonthEarned: earned, MonthSpent: spent,
	}, c)
}

// PointEntries 分页查询积分流水
// @Tags 打卡、积分与通知
// @Summary 分页查询积分流水
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query frontRequest.PointEntriesQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.Page[response.PointEntry],msg=string} "获取成功"
// @Router /miniapp/v1/points/entries [get]
func (*EngagementApi) PointEntries(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var query frontRequest.PointEntriesQuery
	if !apiCommon.BindQuery(c, &query) {
		return
	}
	result, err := engagementService.PointEntries(c.Request.Context(), user.ID, query)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// FeatureUsages 分页查询增强功能调用记录
// @Tags 打卡、积分与通知
// @Summary 分页查询增强功能调用记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query frontRequest.FeatureUsageQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.Page[response.FeatureUsage],msg=string} "获取成功"
// @Router /miniapp/v1/feature-usages [get]
func (*EngagementApi) FeatureUsages(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var query frontRequest.FeatureUsageQuery
	if !apiCommon.BindQuery(c, &query) {
		return
	}
	result, err := engagementService.FeatureUsages(c.Request.Context(), user.ID, query)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// NotificationSummary 获取通知未读数和总数
// @Tags 打卡、积分与通知
// @Summary 获取通知未读数和总数
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Envelope{data=response.NotificationSummaryResult,msg=string} "获取成功"
// @Router /miniapp/v1/notifications/summary [get]
func (*EngagementApi) NotificationSummary(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	unread, total, err := engagementService.NotificationSummary(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.NotificationSummaryResult{
		UnreadCount: unread, Total: total,
	}, c)
}

// Notifications 分页查询通知列表
// @Tags 打卡、积分与通知
// @Summary 分页查询通知列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query frontRequest.NotificationListQuery true "查询条件"
// @Success 200 {object} response.Envelope{data=response.Page[response.Notification],msg=string} "获取成功"
// @Router /miniapp/v1/notifications [get]
func (*EngagementApi) Notifications(c *gin.Context) {
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	var query frontRequest.NotificationListQuery
	if !apiCommon.BindQuery(c, &query) {
		return
	}
	result, err := engagementService.Notifications(c.Request.Context(), user.ID, query)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(result, c)
}

// ReadNotification 将单条通知标记为已读
// @Tags 打卡、积分与通知
// @Summary 将单条通知标记为已读
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Param notificationId path string true "通知ID"
// @Success 200 {object} response.Envelope{data=response.NotificationReadResult,msg=string} "操作成功"
// @Router /miniapp/v1/notifications/{notificationId}/read [post]
func (*EngagementApi) ReadNotification(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	at, err := engagementService.ReadNotification(c.Request.Context(), user.ID, c.Param("notificationId"))
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.NotificationReadResult{Read: true, ReadAt: at}, c)
}

// ReadAllNotifications 将当前用户的全部通知标记为已读
// @Tags 打卡、积分与通知
// @Summary 将当前用户的全部通知标记为已读
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param X-Idempotency-Key header string true "幂等键"
// @Success 200 {object} response.Envelope{data=response.NotificationsReadAllResult,msg=string} "操作成功"
// @Router /miniapp/v1/notifications/read-all [post]
func (*EngagementApi) ReadAllNotifications(c *gin.Context) {
	if _, ok := apiCommon.RequireIdempotencyKey(c); !ok {
		return
	}
	user, ok := apiCommon.CurrentUser(c)
	if !ok {
		return
	}
	affected, err := engagementService.ReadAllNotifications(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	unread, _, err := engagementService.NotificationSummary(c.Request.Context(), user.ID)
	if err != nil {
		apiCommon.Fail(c, err)
		return
	}
	response.OkWithData(response.NotificationsReadAllResult{
		Affected: affected, UnreadCount: unread,
	}, c)
}
