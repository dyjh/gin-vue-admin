package cron_test

import (
	"errors"
	"testing"

	appCron "github.com/dyjh/order-food-mini-app/server/cron"
	"github.com/robfig/cron/v3"
)

type fakeScheduler struct {
	calls []schedulerCall
	err   error
}

type schedulerCall struct {
	cronName string
	spec     string
	taskName string
	options  []cron.Option
	task     func()
}

func (f *fakeScheduler) AddTaskByFunc(cronName string, spec string, task func(), taskName string, option ...cron.Option) (cron.EntryID, error) {
	f.calls = append(f.calls, schedulerCall{
		cronName: cronName, spec: spec, task: task, taskName: taskName, options: option,
	})
	return 1, f.err
}

func TestRegisterAddsOrderFoodBusinessTasks(t *testing.T) {
	scheduler := &fakeScheduler{}

	if err := appCron.Register(scheduler); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if len(scheduler.calls) != 4 {
		t.Fatalf("calls = %d, want 4", len(scheduler.calls))
	}
	deadline := scheduler.calls[0]
	if deadline.cronName != appCron.CronNameMealDeadline ||
		deadline.spec != appCron.SpecMealDeadlineScan ||
		deadline.taskName != appCron.TaskNameMealDeadline {
		t.Fatalf("unexpected deadline task: %+v", deadline)
	}
	refund := scheduler.calls[1]
	if refund.cronName != appCron.CronNameFeatureRefund ||
		refund.spec != appCron.SpecFeatureRefundRetry ||
		refund.taskName != appCron.TaskNameFeatureRefund {
		t.Fatalf("unexpected refund task: %+v", refund)
	}
	subscription := scheduler.calls[2]
	if subscription.cronName != appCron.CronNameSubscriptionDelivery ||
		subscription.spec != appCron.SpecSubscriptionDelivery ||
		subscription.taskName != appCron.TaskNameSubscriptionDelivery {
		t.Fatalf("unexpected subscription task: %+v", subscription)
	}
	preference := scheduler.calls[3]
	if preference.cronName != appCron.CronNamePreferenceEvidence ||
		preference.spec != appCron.SpecPreferenceEvidence ||
		preference.taskName != appCron.TaskNamePreferenceEvidence {
		t.Fatalf("unexpected preference task: %+v", preference)
	}
	for _, call := range scheduler.calls {
		if call.task == nil || len(call.options) != 1 {
			t.Fatalf("invalid registered task: %+v", call)
		}
	}
}

func TestRegisterReturnsSchedulerError(t *testing.T) {
	wantErr := errors.New("add task failed")
	scheduler := &fakeScheduler{err: wantErr}

	err := appCron.Register(scheduler)

	if !errors.Is(err, wantErr) {
		t.Fatalf("Register() error = %v, want %v", err, wantErr)
	}
}
