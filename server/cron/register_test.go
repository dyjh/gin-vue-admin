package cron_test

import (
	"errors"
	"testing"

	appCron "github.com/dyjh/order-food-mini-app/server/cron"
	"github.com/robfig/cron/v3"
)

type fakeScheduler struct {
	cronName string
	spec     string
	taskName string
	options  []cron.Option
	task     func()
	err      error
}

func (f *fakeScheduler) AddTaskByFunc(cronName string, spec string, task func(), taskName string, option ...cron.Option) (cron.EntryID, error) {
	f.cronName = cronName
	f.spec = spec
	f.task = task
	f.taskName = taskName
	f.options = option
	return 1, f.err
}

func TestRegisterAddsCancelExpiredOrderTask(t *testing.T) {
	scheduler := &fakeScheduler{}

	if err := appCron.Register(scheduler); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if scheduler.cronName != appCron.CronNameOrder {
		t.Fatalf("cronName = %q, want %q", scheduler.cronName, appCron.CronNameOrder)
	}
	if scheduler.spec != appCron.SpecCancelExpiredOrders {
		t.Fatalf("spec = %q, want %q", scheduler.spec, appCron.SpecCancelExpiredOrders)
	}
	if scheduler.taskName != appCron.TaskNameCancelExpiredOrders {
		t.Fatalf("taskName = %q, want %q", scheduler.taskName, appCron.TaskNameCancelExpiredOrders)
	}
	if scheduler.task == nil {
		t.Fatal("task is nil")
	}
	if len(scheduler.options) != 1 {
		t.Fatalf("len(options) = %d, want 1", len(scheduler.options))
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
