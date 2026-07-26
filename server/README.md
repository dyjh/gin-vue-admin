## server项目结构

```shell
├── api
│   └── v1
├── config
├── core
├── docs
├── global
├── initialize
│   └── internal
├── middleware
├── model
│   ├── request
│   └── response
├── packfile
├── resource
│   ├── excel
│   ├── page
│   └── template
├── router
├── service
├── source
└── utils
    ├── timer
    └── upload
```

| 文件夹       | 说明                    | 描述                        |
| ------------ | ----------------------- | --------------------------- |
| `api`        | api层                   | api层 |
| `--v1`       | v1版本接口              | v1版本接口                  |
| `config`     | 配置包                  | config.yaml对应的配置结构体 |
| `core`       | 核心文件                | 核心组件(zap, viper, server)的初始化 |
| `docs`       | swagger文档目录         | swagger文档目录 |
| `global`     | 全局对象                | 全局对象 |
| `initialize` | 初始化 | router,redis,gorm,validator, timer的初始化 |
| `--internal` | 初始化内部函数 | gorm 的 longger 自定义,在此文件夹的函数只能由 `initialize` 层进行调用 |
| `middleware` | 中间件层 | 用于存放 `gin` 中间件代码 |
| `model`      | 模型层                  | 模型对应数据表              |
| `--request`  | 入参结构体              | 接收前端发送到后端的数据。  |
| `--response` | 出参结构体              | 返回给前端的数据结构体      |
| `packfile`   | 静态文件打包            | 静态文件打包 |
| `resource`   | 静态资源文件夹          | 负责存放静态文件                |
| `--excel` | excel导入导出默认路径 | excel导入导出默认路径 |
| `--page` | 表单生成器 | 表单生成器 打包后的dist |
| `--template` | 模板 | 模板文件夹,存放的是代码生成器的模板 |
| `router`     | 路由层                  | 路由层 |
| `service`    | service层               | 存放业务逻辑问题 |
| `source` | source层 | 存放初始化数据的函数 |
| `utils`      | 工具包                  | 工具函数封装            |
| `--timer` | timer | 定时器接口封装 |
| `--upload`      | oss                  | oss接口封装        |

## orderfood 数据库测试

`orderfood` 的 GORM 与数据库集成测试使用 MySQL 8，测试代码不会读取 `config.yaml`，也不会回退到 SQLite。根目录提供了独立测试实例：

```shell
make test-orderfood-mysql-up
make test-orderfood-mysql
make test-orderfood-mysql-down
```

测试默认连接本机 `13307` 端口。若使用已有 MySQL 测试实例，可覆盖 `ORDERFOOD_TEST_MYSQL_DSN`；DSN 中的基础数据库名必须以 `_test` 结尾，连接账号需要拥有测试实例内创建和删除数据库的权限。每个测试会创建独立数据库并在结束时自动删除，不会复用基础数据库中的表。

## orderfood 微信登录与订阅消息

小程序登录和微信订阅消息投递共用管理端一级菜单“微信配置”中当前生效的 AppID 和 AppSecret。AppSecret 加密保存且不回显；首次配置和更换 AppID 时必须填写，后续留空表示保留当前值。

微信 AppSecret、`openid`、`unionid` 和 `session_key` 使用 32 字节密钥加密后落库。密钥只从 `config.yaml` 的 `orderfood.identity-key` 读取，必须是 32 字节密钥的标准 Base64 编码：

```yaml
orderfood:
    identity-key: "32字节密钥的Base64编码"
```

可使用 `openssl rand -base64 32` 生成一次，写入后需要重启服务。配置投入使用后必须固定保存，不能在服务重启或发布时更换，否则已有 AppSecret 和微信用户身份数据将无法解密。该字段不会通过 GVA 系统配置接口返回，保存其他系统配置时也不会覆盖它。

服务端通过 `jscode2session` 兑换 `wx.login` 的一次性 code；微信 AccessToken 只缓存在进程内存中。发送订阅消息遇到微信判定令牌无效或过期时，会清除对应旧令牌、重新获取一次并补发一次。

## orderfood 阿里云图片审核

管理端“图片审核配置”保存后立即生效，不使用草稿、发布或历史版本。第一版凭据引用只支持 `env://环境变量名`，数据库、接口响应和日志都不保存或回显 AccessKey 明文。

例如管理端填写：

```text
env://ORDERFOOD_ALIYUN_MODERATION
```

对应环境变量内容：

```json
{"accessKeyId":"RAM_ACCESS_KEY_ID","accessKeySecret":"RAM_ACCESS_KEY_SECRET"}
```

使用临时 STS 凭据时可额外提供 `securityToken`。Endpoint 必须是无路径、查询参数和用户信息的 HTTPS 地址，例如 `https://green-cip.cn-shanghai.aliyuncs.com`。

生产调用使用阿里云官方 `green-20220302` SDK。连接测试会获取临时 OSS 上传令牌；图片审核会先把本地图片上传到阿里云提供的临时 OSS 空间，再调用同步 `ImageModeration`，并尽力清理临时对象。供应商只要返回风险标签、请求超时、鉴权失败、限流或异常，用户上传都不会被当作审核通过。

AI 供应商密钥引用同样只支持 `env://环境变量名`，但对应环境变量内容直接是该供应商的 API Key 字符串，不是上述阿里云图片审核 JSON。后台保存的仍然只是引用，密钥不会进入接口响应。
