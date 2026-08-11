package common

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	commonModel "github.com/dyjh/order-food-mini-app/server/model/common"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const adminIdempotencyTTL = 24 * time.Hour

// IdempotentAction 定义需要在幂等事务中执行并缓存结果的业务操作。
type IdempotentAction func(tx *gorm.DB) (interface{}, error)

// IdempotencyService 提供后台幂等执行业务能力。
type IdempotencyService struct {
	DB  *gorm.DB         // 业务数据库
	Now func() time.Time // 可注入的当前时间函数
}

// NewIdempotencyService 创建幂等服务实例。
func NewIdempotencyService(db *gorm.DB) *IdempotencyService {
	return &IdempotencyService{DB: db, Now: time.Now}
}

func (service *IdempotencyService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *IdempotencyService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// HashIdempotencyPayload 计算幂等请求负载摘要。
func HashIdempotencyPayload(payload interface{}) (string, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", appErrors.AdminBadRequest.Wrap(err, "encode idempotency payload")
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

// Execute runs the mutation and marks the idempotency row completed in one
// transaction. A failed mutation rolls the processing row back so a corrected
// request may be retried with a new or unchanged key.
func (service *IdempotencyService) Execute(
	ctx context.Context,
	administratorID uint,
	endpointID string,
	idempotencyKey string,
	payload interface{},
	action IdempotentAction,
) (json.RawMessage, bool, error) {
	if administratorID == 0 || endpointID == "" || idempotencyKey == "" || action == nil {
		return nil, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return nil, false, appErrors.AdminInternal.DefaultMsg()
	}
	requestHash, err := HashIdempotencyPayload(payload)
	if err != nil {
		return nil, false, err
	}

	var responseJSON json.RawMessage
	var replayed bool
	now := service.now()
	// 幂等记录与业务操作共用事务：相同请求直接回放，载荷不同则拒绝，任何业务失败都不会缓存成功结果。
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record commonModel.AdminIdempotencyRecord
		findErr := tx.
			Where(
				"administrator_id = ? AND endpoint_id = ? AND idempotency_key = ?",
				administratorID,
				endpointID,
				idempotencyKey,
			).
			First(&record).Error
		switch {
		case findErr == nil && record.ExpiresAt.After(now):
			if record.RequestHash != requestHash {
				return appErrors.AdminIdempotencyConflict.DefaultMsg()
			}
			switch record.State {
			case commonModel.AdminIdempotencyProcessing:
				return appErrors.AdminRequestProcessing.DefaultMsg()
			case commonModel.AdminIdempotencyCompleted:
				if len(record.ResponseJSON) == 0 {
					return appErrors.AdminInternal.DefaultMsg()
				}
				responseJSON = append(json.RawMessage(nil), record.ResponseJSON...)
				replayed = true
				return nil
			default:
				return appErrors.AdminInternal.DefaultMsg()
			}
		case findErr == nil:
			if deleteErr := tx.Delete(&record).Error; deleteErr != nil {
				return appErrors.AdminInternal.Wrap(deleteErr, "delete expired idempotency record")
			}
		case findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound):
			return appErrors.AdminInternal.Wrap(findErr, "query idempotency record")
		}

		record = commonModel.AdminIdempotencyRecord{
			AdministratorID: administratorID,
			EndpointID:      endpointID,
			IdempotencyKey:  idempotencyKey,
			RequestHash:     requestHash,
			State:           commonModel.AdminIdempotencyProcessing,
			CreatedAt:       now,
			UpdatedAt:       now,
			ExpiresAt:       now.Add(adminIdempotencyTTL),
		}
		if createErr := tx.Create(&record).Error; createErr != nil {
			return appErrors.AdminRequestProcessing.Wrap(createErr, "create idempotency record")
		}

		result, actionErr := action(tx)
		if actionErr != nil {
			return actionErr
		}
		encoded, encodeErr := json.Marshal(result)
		if encodeErr != nil {
			return appErrors.AdminInternal.Wrap(encodeErr, "encode idempotent response")
		}
		responseJSON = encoded
		update := tx.Model(&record).Updates(map[string]interface{}{
			"state":         commonModel.AdminIdempotencyCompleted,
			"response_code": 0,
			"response_json": datatypes.JSON(encoded),
			"updated_at":    now,
		})
		if update.Error != nil || update.RowsAffected != 1 {
			if update.Error != nil {
				return appErrors.AdminInternal.Wrap(update.Error, "complete idempotency record")
			}
			return appErrors.AdminInternal.DefaultMsg()
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return responseJSON, replayed, nil
}
