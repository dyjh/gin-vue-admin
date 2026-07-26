# Model 示例

## 这个文件负责什么

Model 负责定义数据库实体与持久化字段，是 Service 和数据库交互的基础。

## 什么时候应该这样写

- 新增一张业务表
- 为现有表补字段
- 需要定义 GORM 结构与关联关系

## 推荐写法示例

```go
package system

import (
	"time"

	"github.com/dyjh/order-food-mini-app/server/global"
	"gorm.io/datatypes"
)

// EmailPush 表示邮件推送任务。
type EmailPush struct {
	global.GVA_MODEL // GVA基础模型字段
	Subject    string         `json:"subject" form:"subject" gorm:"column:subject;type:varchar(255);not null;comment:邮件主题;"`                       // 邮件主题
	Content    string         `json:"content" form:"content" gorm:"column:content;type:longtext;not null;comment:邮件正文;"`                          // 邮件正文
	SendType   uint           `json:"sendType" form:"sendType" gorm:"column:send_type;type:tinyint unsigned;not null;default:1;comment:发送类型;"`      // 发送类型：1立即发送，2定时发送
	SendTime   *time.Time     `json:"sendTime" form:"sendTime" gorm:"column:send_time;type:datetime;default:null;comment:定时发送时间;"`                // 定时发送时间
	SendUsers  datatypes.JSON `json:"sendUsers" form:"sendUsers" gorm:"column:send_users;type:json;default:null;comment:发送用户;"`                    // 发送用户；空值表示全部用户
	SendStatus uint           `json:"sendStatus" form:"sendStatus" gorm:"column:send_status;type:tinyint unsigned;not null;default:1;comment:发送状态;"` // 发送状态：1待发送，2已发送
	Remark     string         `json:"remark" form:"remark" gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注;"`                   // 备注
}

// TableName 指定EmailPush对应的数据表名。
func (EmailPush) TableName() string {
	return "sys_email_push"
}
```

## 为什么这样写

- 继承 `global.GVA_MODEL`，保持主键和时间字段风格一致
- `json` 标签用于接口输出
- `gorm` 标签明确列名、数据库类型、空值或默认值、索引和数据库注释
- 字段命名尽量清晰、稳定，便于前后端保持一致
- 字段说明与字段保持同行，方便审查标签和业务语义
- 表名使用短前缀并控制长度，避免生成难维护的长索引名

## 常见错误

- 缺少 `json` 或 `gorm` 标签，或只写 `size` 而没有明确列名、数据库类型和注释
- 表名或索引名使用冗长的业务全称
- 结构体字段没有同行说明
- 把仅用于请求或展示的字段直接写入数据库 model
- 同一个字段在前后端使用不同类型
- 忽略 `Status`、`ID`、时间字段这类高风险类型一致性问题

## 真实参考文件

- `server/model/system/sys_api_token.go`
- `server/model/system/sys_user.go`
