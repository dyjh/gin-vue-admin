package middleware

import (
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/front/service"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ContextUserID = "orderfood_front_user_id"
	ContextUser   = "orderfood_front_user"
)

type AuthMiddleware struct {
	Service *service.AuthService // 登录认证服务
	DB      *gorm.DB             // 数据库连接
	Now     func() time.Time     // 当前时间函数
}

func (middleware AuthMiddleware) authService() *service.AuthService {
	if middleware.Service != nil {
		return middleware.Service
	}
	return &service.ServiceGroupApp.AuthService
}

// database 返回中间件使用的数据库连接。
func (middleware AuthMiddleware) database() *gorm.DB {
	if middleware.DB != nil {
		return middleware.DB
	}
	return global.GVA_DB
}

// now 返回中间件使用的UTC时间。
func (middleware AuthMiddleware) now() time.Time {
	if middleware.Now != nil {
		return middleware.Now().UTC()
	}
	return time.Now().UTC()
}

// Handle 校验小程序访问令牌，并在成功业务请求后登记用户活跃自然日。
func (middleware AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.Fields(header)
		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			_ = c.Error(appErrors.FrontLoginExpired.DefaultMsg())
			c.Abort()
			return
		}
		user, err := middleware.authService().Authenticate(c.Request.Context(), parts[1])
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Set(ContextUserID, user.ID)
		c.Set(ContextUser, user)
		c.Next()
		if len(c.Errors) > 0 || c.Writer.Status() < 200 || c.Writer.Status() >= 400 {
			return
		}
		// 活跃记录按用户和上海自然日去重，不延长请求链路之外的状态语义。
		now := middleware.now()
		location, locationErr := time.LoadLocation("Asia/Shanghai")
		if locationErr != nil {
			location = time.FixedZone("Asia/Shanghai", 8*60*60)
		}
		if db := middleware.database(); db != nil {
			_ = db.WithContext(c.Request.Context()).
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(&orderfoodModel.FrontUserActivityDay{
					UserID: user.ID, ActiveDate: now.In(location).Format("2006-01-02"),
					FirstAt: now, CreatedAt: now,
				}).Error
		}
	}
}

// CurrentUser 从Gin上下文读取当前小程序用户。
func CurrentUser(c *gin.Context) (orderfoodModel.MiniAppUser, bool) {
	value, exists := c.Get(ContextUser)
	if !exists {
		return orderfoodModel.MiniAppUser{}, false
	}
	user, ok := value.(orderfoodModel.MiniAppUser)
	return user, ok
}
