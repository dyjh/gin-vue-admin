package cron

import (
	"context"
	"errors"
	"fmt"

	frontService "github.com/dyjh/order-food-mini-app/server/front/service"
	robfigcron "github.com/robfig/cron/v3"
)

const (
	// CronNameMealDeadline 标识饭局截止时间扫描任务。
	CronNameMealDeadline = "OrderFoodMealDeadline"
	// SpecMealDeadlineScan 定义饭局截止时间扫描频率。
	SpecMealDeadlineScan = "0 */1 * * * *"
	// TaskNameMealDeadline 描述饭局截止时间扫描任务。
	TaskNameMealDeadline = "关闭达到点单截止时间的饭局"
	// CronNameFeatureRefund 标识增强功能待退积分补偿任务。
	CronNameFeatureRefund = "OrderFoodFeatureRefund"
	// SpecFeatureRefundRetry 定义待退积分补偿频率。
	SpecFeatureRefundRetry = "*/30 * * * * *"
	// TaskNameFeatureRefund 描述待退积分补偿任务。
	TaskNameFeatureRefund = "重试增强功能待退积分"
	// CronNameSubscriptionDelivery 标识微信订阅消息投递任务。
	CronNameSubscriptionDelivery = "OrderFoodSubscriptionDelivery"
	// SpecSubscriptionDelivery 定义微信订阅消息投递频率。
	SpecSubscriptionDelivery = "*/5 * * * * *"
	// TaskNameSubscriptionDelivery 描述微信订阅消息投递任务。
	TaskNameSubscriptionDelivery = "投递饭局最终结果订阅消息"
	// CronNamePreferenceEvidence 标识用户偏好证据处理任务。
	CronNamePreferenceEvidence = "OrderFoodPreferenceEvidence"
	// SpecPreferenceEvidence 定义用户偏好证据处理频率。
	SpecPreferenceEvidence = "*/5 * * * * *"
	// TaskNamePreferenceEvidence 描述用户偏好证据处理任务。
	TaskNamePreferenceEvidence = "分析并聚合用户偏好证据"
)

// ErrNilScheduler 表示没有可用的统一定时任务调度器。
var ErrNilScheduler = errors.New("cron scheduler is nil")

// Scheduler 定义项目业务定时任务注册所需的最小调度能力。
type Scheduler interface {
	AddTaskByFunc(cronName string, spec string, task func(), taskName string, option ...robfigcron.Option) (robfigcron.EntryID, error)
}

// Register 将饭局截止、积分退款、订阅投递和偏好聚合任务注册到统一调度器。
func Register(scheduler Scheduler) error {
	if scheduler == nil {
		return ErrNilScheduler
	}

	// option := []robfigcron.Option{robfigcron.WithSeconds()}
	// if _, err := scheduler.AddTaskByFunc(
	// 	CronNameMealDeadline,
	// 	SpecMealDeadlineScan,
	// 	processExpiredMeals,
	// 	TaskNameMealDeadline,
	// 	option...,
	// ); err != nil {
	// 	return err
	// }
	// if _, err := scheduler.AddTaskByFunc(
	// 	CronNameFeatureRefund,
	// 	SpecFeatureRefundRetry,
	// 	processPendingFeatureRefunds,
	// 	TaskNameFeatureRefund,
	// 	option...,
	// ); err != nil {
	// 	return err
	// }
	// if _, err := scheduler.AddTaskByFunc(
	// 	CronNameSubscriptionDelivery,
	// 	SpecSubscriptionDelivery,
	// 	processPendingSubscriptionMessages,
	// 	TaskNameSubscriptionDelivery,
	// 	option...,
	// ); err != nil {
	// 	return err
	// }
	// _, err := scheduler.AddTaskByFunc(
	// 	CronNamePreferenceEvidence,
	// 	SpecPreferenceEvidence,
	// 	processPendingPreferenceEvidence,
	// 	TaskNamePreferenceEvidence,
	// 	option...,
	// )
	// return err
	return nil
}

// processExpiredMeals 扫描到期饭局；错误仅记录并等待下一轮幂等重试。
func processExpiredMeals() {
	if err := frontService.ServiceGroupApp.MealService.
		ProcessExpiredMeals(context.Background(), 100); err != nil {
		fmt.Println("process expired order food meals error:", err)
	}
}

// processPendingFeatureRefunds 重试待退积分；错误仅记录并等待下一轮幂等重试。
func processPendingFeatureRefunds() {
	if err := frontService.ServiceGroupApp.EngagementService.
		ProcessPendingRefunds(context.Background(), 100); err != nil {
		fmt.Println("process order food feature refunds error:", err)
	}
}

// processPendingSubscriptionMessages 投递待发送消息；错误仅记录并等待下一轮幂等重试。
func processPendingSubscriptionMessages() {
	if err := frontService.ServiceGroupApp.SubscriptionDeliveryService.
		ProcessPendingMessages(context.Background(), 100); err != nil {
		fmt.Println("process order food subscription messages error:", err)
	}
}

// processPendingPreferenceEvidence 分析并聚合偏好证据；错误仅记录并等待下一轮重试。
func processPendingPreferenceEvidence() {
	if err := frontService.ServiceGroupApp.PreferenceService.
		ProcessPendingEvidence(context.Background(), 100); err != nil {
		fmt.Println("process order food preference evidence error:", err)
	}
}
