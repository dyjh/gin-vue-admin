package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	"github.com/dyjh/order-food-mini-app/server/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type captureWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (writer *captureWriter) Write(data []byte) (int, error) {
	writer.body.Write(data)
	return writer.ResponseWriter.Write(data)
}

func (writer *captureWriter) WriteString(value string) (int, error) {
	writer.body.WriteString(value)
	return writer.ResponseWriter.WriteString(value)
}

type IdempotencyMiddleware struct {
	DB  *gorm.DB
	Now func() time.Time
}

func (middleware IdempotencyMiddleware) database() *gorm.DB {
	if middleware.DB != nil {
		return middleware.DB
	}
	return global.GVA_DB
}

func (middleware IdempotencyMiddleware) now() time.Time {
	if middleware.Now != nil {
		return middleware.Now().UTC()
	}
	return time.Now().UTC()
}

func mutationMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
}

func requestDigest(c *gin.Context, body []byte) string {
	sum := sha256.New()
	sum.Write([]byte(c.Request.Method))
	sum.Write([]byte{0})
	sum.Write([]byte(c.FullPath()))
	sum.Write([]byte{0})
	sum.Write([]byte(c.Request.URL.RawQuery))
	sum.Write([]byte{0})
	sum.Write(body)
	return hex.EncodeToString(sum.Sum(nil))
}

func (middleware IdempotencyMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mutationMethod(c.Request.Method) ||
			c.FullPath() == "/api/miniapp/v1/uploads/images" {
			c.Next()
			return
		}
		key := strings.TrimSpace(c.GetHeader("X-Idempotency-Key"))
		if key == "" || len(key) > 128 {
			_ = c.Error(appErrors.FrontBadRequest.DefaultMsg())
			c.Abort()
			return
		}
		// 将已校验的幂等键传入 Service 上下文，供 AI 调用记录关联完整请求链路。
		c.Request = c.Request.WithContext(utils.WithIdempotencyKey(c.Request.Context(), key))
		userIDValue, exists := c.Get(ContextUserID)
		userID, valid := userIDValue.(string)
		if !exists || !valid || userID == "" {
			_ = c.Error(appErrors.FrontLoginExpired.DefaultMsg())
			c.Abort()
			return
		}
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, 2<<20))
		if err != nil {
			_ = c.Error(appErrors.FrontBadRequest.DefaultMsg())
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		hash := requestDigest(c, body)
		endpoint := c.Request.Method + " " + c.FullPath()
		now := middleware.now()
		record := orderfoodModel.FrontIdempotencyRecord{
			UserID: userID, Endpoint: endpoint, Key: key, RequestHash: hash,
			State: "processing", CreatedAt: now, UpdatedAt: now, ExpiresAt: now.Add(24 * time.Hour),
		}
		create := middleware.database().WithContext(c.Request.Context()).
			Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
		if create.Error != nil {
			_ = c.Error(appErrors.FrontInternal.Wrap(create.Error, "reserve front idempotency key"))
			c.Abort()
			return
		}
		if create.RowsAffected == 0 {
			var existing orderfoodModel.FrontIdempotencyRecord
			if err := middleware.database().WithContext(c.Request.Context()).
				First(&existing, "user_id = ? AND endpoint = ? AND key = ?", userID, endpoint, key).Error; err != nil {
				_ = c.Error(appErrors.FrontInternal.Wrap(err, "load front idempotency key"))
				c.Abort()
				return
			}
			if existing.RequestHash != hash {
				_ = c.Error(appErrors.FrontIdempotencyConflict.DefaultMsg())
				c.Abort()
				return
			}
			if existing.State == "completed" {
				c.Data(existing.ResponseStatus, "application/json; charset=utf-8", existing.ResponseBody)
				c.Abort()
				return
			}
			_ = c.Error(appErrors.FrontRequestProcessing.DefaultMsg())
			c.Abort()
			return
		}

		writer := &captureWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()
		if len(c.Errors) > 0 || c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
			_ = middleware.database().WithContext(c.Request.Context()).Delete(&record).Error
			return
		}
		if err := middleware.database().WithContext(c.Request.Context()).
			Model(&orderfoodModel.FrontIdempotencyRecord{}).
			Where("id = ?", record.ID).
			Updates(map[string]interface{}{
				"state": "completed", "response_status": c.Writer.Status(),
				"response_body": append([]byte(nil), writer.body.Bytes()...), "updated_at": middleware.now(),
			}).Error; err != nil {
			_ = c.Error(appErrors.FrontInternal.Wrap(err, "complete front idempotency key"))
		}
	}
}
