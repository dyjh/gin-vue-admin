package cron

import (
	"errors"

	robfigcron "github.com/robfig/cron/v3"
)

const (
	CronNameOrder               = "OrderCron"
	SpecCancelExpiredOrders     = "0 */1 * * * *"
	TaskNameCancelExpiredOrders = "取消超时未支付订单"
)

var ErrNilScheduler = errors.New("cron scheduler is nil")

type Scheduler interface {
	AddTaskByFunc(cronName string, spec string, task func(), taskName string, option ...robfigcron.Option) (robfigcron.EntryID, error)
}

func Register(scheduler Scheduler) error {
	if scheduler == nil {
		return ErrNilScheduler
	}

	option := []robfigcron.Option{robfigcron.WithSeconds()}
	_, err := scheduler.AddTaskByFunc(
		CronNameOrder,
		SpecCancelExpiredOrders,
		cancelExpiredOrders,
		TaskNameCancelExpiredOrders,
		option...,
	)
	return err
}
