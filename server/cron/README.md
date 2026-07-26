# Cron 定时任务

`server/cron` 用于存放项目级业务定时任务。

项目已经通过 `global.GVA_Timer` 使用 `github.com/robfig/cron/v3`。除非任务有独立生命周期要求，
业务定时任务不要再自行创建新的 cron 实例。

## 目录结构

```text
server/cron/
  register.go  # 统一注册项目 cron 任务
```

## 注册方式

`initialize.Timer()` 只调用一次 `cron.Register(global.GVA_Timer)`。新增任务时，在 `Register`
中注册任务，具体业务处理保留在对应领域服务中。

当前使用 `cron.WithSeconds()`，cron 表达式按秒级格式编写。

当前任务包括饭局点单截止、失败积分补偿、微信订阅消息投递和偏好证据聚合。
