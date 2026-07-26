package initialize

import (
	"context"
	"fmt"

	appCron "github.com/dyjh/order-food-mini-app/server/cron"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/dyjh/order-food-mini-app/server/task"

	"github.com/robfig/cron/v3"

	"github.com/dyjh/order-food-mini-app/server/global"
)

// Timer 注册并异步启动系统与来干饭业务定时任务。
func Timer() {
	go func() {
		var option []cron.Option
		option = append(option, cron.WithSeconds())
		// 清理DB定时任务
		_, err := global.GVA_Timer.AddTaskByFunc("ClearDB", "@daily", func() {
			err := task.ClearTable(global.GVA_DB) // 定时任务方法定在task文件包中
			if err != nil {
				fmt.Println("timer error:", err)
			}
		}, "定时清理数据库【日志，黑名单】内容", option...)
		if err != nil {
			fmt.Println("add timer error:", err)
		}

		// 复制链治理任务分批执行，失败项保留给管理员在治理记录页重试。
		_, err = global.GVA_Timer.AddTaskByFunc("OrderFoodGovernanceJobs", "@every 5s", func() {
			if processErr := orderfoodService.ServiceGroupApp.Governance.
				ProcessPendingJobs(context.Background(), 100); processErr != nil {
				fmt.Println("process order food governance jobs error:", processErr)
			}
		}, "分批处理来干饭严重违规复制链任务", option...)
		if err != nil {
			fmt.Println("add order food governance timer error:", err)
		}

		if err := appCron.Register(global.GVA_Timer); err != nil {
			fmt.Println("add cron error:", err)
		}

		// 其他定时任务定在这里 参考上方使用方法

		//_, err := global.GVA_Timer.AddTaskByFunc("定时任务标识", "corn表达式", func() {
		//	具体执行内容...
		//  ......
		//}, option...)
		//if err != nil {
		//	fmt.Println("add timer error:", err)
		//}
	}()
}
