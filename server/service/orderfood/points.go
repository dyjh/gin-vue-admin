package orderfood

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	orderfoodRequest "github.com/dyjh/order-food-mini-app/server/model/orderfood/request"
	orderfoodResponse "github.com/dyjh/order-food-mini-app/server/model/orderfood/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// PermissionPointsRead 表示读取积分流水所需的权限。
	PermissionPointsRead = "orderfood:points:read"
	// PermissionPointsAdjust 表示人工调整积分所需的权限。
	PermissionPointsAdjust    = "orderfood:points:adjust"
	PermissionPointRuleRead   = "orderfood:point-rule:read"
	PermissionPointRuleUpdate = "orderfood:point-rule:update"

	DefaultDailyCheckinReward int64 = 1
	pointRuleSingletonKey           = "platform"
	pointAdjustmentPreviewTTL       = 10 * time.Minute
)

// PointsService 提供积分规则业务能力。
type PointsService struct {
	DB            *gorm.DB            // 业务数据库
	Permission    *PermissionService  // GVA权限服务
	Idempotency   *IdempotencyService // 幂等执行服务
	MutationAudit MutationAuditWriter // 配置变更审计写入器
	Now           func() time.Time    // 可注入的当前时间函数
}

// NewPointsService 创建积分服务实例。
func NewPointsService(
	db *gorm.DB,
	permission *PermissionService,
	idempotency *IdempotencyService,
) *PointsService {
	if permission == nil {
		permission = NewPermissionService(db)
	}
	if idempotency == nil {
		idempotency = NewIdempotencyService(db)
	}
	return &PointsService{
		DB:            db,
		Permission:    permission,
		Idempotency:   idempotency,
		MutationAudit: gormMutationAuditWriter{},
		Now:           time.Now,
	}
}

func (service *PointsService) database() *gorm.DB {
	if service != nil && service.DB != nil {
		return service.DB
	}
	return global.GVA_DB
}

func (service *PointsService) now() time.Time {
	if service != nil && service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

// pointEntryResponse 将积分流水模型转换为管理端响应。
func pointEntryResponse(
	row orderfoodModel.FrontPointEntry,
	user orderfoodModel.MiniAppUser,
) orderfoodResponse.PointEntry {
	var administrator *orderfoodResponse.AdministratorSummary
	if row.AdministratorID != nil && row.AdministratorUsername != nil {
		administrator = &orderfoodResponse.AdministratorSummary{
			ID:       strconv.FormatUint(uint64(*row.AdministratorID), 10),
			Username: *row.AdministratorUsername,
			Nickname: row.AdministratorNickname,
		}
	}
	return orderfoodResponse.PointEntry{
		ID:                row.ID,
		User:              subscriptionUserReference(user),
		Type:              string(row.Type),
		Scene:             row.Scene,
		Title:             row.Title,
		Description:       row.Description,
		Amount:            row.Amount,
		BalanceAfter:      row.BalanceAfter,
		RelatedObjectType: row.RelatedObjectType,
		RelatedObjectID:   row.RelatedObjectID,
		RelatedEntryID:    row.RelatedEntryID,
		RequestID:         row.RequestID,
		IdempotencyKey:    row.IdempotencyKey,
		Administrator:     administrator,
		CreatedAt:         row.CreatedAt,
	}
}

// ListPointEntries 分页查询不可变的积分流水。
func (service *PointsService) ListPointEntries(
	ctx context.Context,
	query orderfoodRequest.PointEntryListQuery,
) (orderfoodResponse.Page[orderfoodResponse.PointEntry], error) {
	query.ApplyDefaults()
	sortColumns := map[string]string{
		"createdAt": "created_at",
		"amount":    "amount",
	}
	column, ok := sortColumns[query.SortBy]
	if !ok {
		return orderfoodResponse.Page[orderfoodResponse.PointEntry]{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return orderfoodResponse.Page[orderfoodResponse.PointEntry]{}, appErrors.AdminInternal.DefaultMsg()
	}
	statement := db.WithContext(ctx).Model(&orderfoodModel.FrontPointEntry{})
	if value := strings.TrimSpace(query.UserID); value != "" {
		statement = statement.Where("user_id = ?", value)
	}
	if query.Type != "" {
		statement = statement.Where("type = ?", query.Type)
	}
	if value := strings.TrimSpace(query.Scene); value != "" {
		statement = statement.Where("scene = ?", value)
	}
	if value := strings.TrimSpace(query.RelatedObjectType); value != "" {
		statement = statement.Where("related_object_type = ?", value)
	}
	if value := strings.TrimSpace(query.RelatedObjectID); value != "" {
		statement = statement.Where("related_object_id = ?", value)
	}
	if value := strings.TrimSpace(query.RequestID); value != "" {
		statement = statement.Where("request_id = ?", value)
	}
	if value := strings.TrimSpace(query.IdempotencyKey); value != "" {
		statement = statement.Where("idempotency_key = ?", value)
	}
	if query.CreatedFrom != nil {
		statement = statement.Where("created_at >= ?", *query.CreatedFrom)
	}
	if query.CreatedTo != nil {
		statement = statement.Where("created_at <= ?", *query.CreatedTo)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.PointEntry]{}, appErrors.AdminInternal.Wrap(err, "count point entries")
	}
	rows := make([]orderfoodModel.FrontPointEntry, 0)
	if err := statement.Order(column + " " + query.SortOrder).
		Order("id asc").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&rows).Error; err != nil {
		return orderfoodResponse.Page[orderfoodResponse.PointEntry]{}, appErrors.AdminInternal.Wrap(err, "list point entries")
	}
	userIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID)
	}
	users := make([]orderfoodModel.MiniAppUser, 0)
	if len(userIDs) > 0 {
		if err := db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return orderfoodResponse.Page[orderfoodResponse.PointEntry]{}, appErrors.AdminInternal.Wrap(err, "load point entry users")
		}
	}
	userMap := make(map[string]orderfoodModel.MiniAppUser, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}
	list := make([]orderfoodResponse.PointEntry, 0, len(rows))
	for _, row := range rows {
		list = append(list, pointEntryResponse(row, userMap[row.UserID]))
	}
	return orderfoodResponse.Page[orderfoodResponse.PointEntry]{
		Page: query.Page, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// pointAdjustmentPreviewClaims 表示积分调整预览令牌中的安全声明。
type pointAdjustmentPreviewClaims struct {
	RequestHash         string `json:"requestHash"`         // 预览请求摘要
	ExpectedUserVersion int64  `json:"expectedUserVersion"` // 预览时用户版本
	BalanceBefore       int64  `json:"balanceBefore"`       // 预览时积分余额
	ExpiresAt           int64  `json:"expiresAt"`           // 过期时间戳
}

// pointAdjustmentPreviewSecret 返回积分调整预览令牌的签名密钥。
func (service *PointsService) pointAdjustmentPreviewSecret() []byte {
	signingKey := strings.TrimSpace(global.GVA_CONFIG.JWT.SigningKey)
	if signingKey == "" {
		return nil
	}
	sum := sha256.Sum256([]byte("orderfood-point-adjustment-preview:" + signingKey))
	return sum[:]
}

// pointAdjustmentRequestHash 计算绑定管理员与调整参数的预览摘要。
func pointAdjustmentRequestHash(
	administratorID uint,
	input orderfoodRequest.PointAdjustmentPreviewInput,
) (string, error) {
	payload := struct {
		AdministratorID uint   `json:"administratorId"`
		UserID          string `json:"userId"`
		Direction       string `json:"direction"`
		Amount          int64  `json:"amount"`
		Reason          string `json:"reason"`
	}{
		AdministratorID: administratorID,
		UserID:          strings.TrimSpace(input.UserID),
		Direction:       input.Direction,
		Amount:          input.Amount,
		Reason:          strings.TrimSpace(input.Reason),
	}
	return HashIdempotencyPayload(payload)
}

// signPointAdjustmentPreview 签发十分钟有效的积分调整预览令牌。
func (service *PointsService) signPointAdjustmentPreview(
	requestHash string,
	expectedUserVersion int64,
	balanceBefore int64,
	expiresAt time.Time,
) (string, error) {
	secret := service.pointAdjustmentPreviewSecret()
	if len(secret) == 0 {
		return "", appErrors.AdminInternal.DefaultMsg()
	}
	claims, err := json.Marshal(pointAdjustmentPreviewClaims{
		RequestHash:         requestHash,
		ExpectedUserVersion: expectedUserVersion,
		BalanceBefore:       balanceBefore,
		ExpiresAt:           expiresAt.Unix(),
	})
	if err != nil {
		return "", appErrors.AdminInternal.Wrap(err, "encode point adjustment preview")
	}
	payload := base64.RawURLEncoding.EncodeToString(claims)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(payload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + signature, nil
}

// verifyPointAdjustmentPreview 校验预览令牌签名、有效期和请求绑定关系。
func (service *PointsService) verifyPointAdjustmentPreview(
	token string,
	expectedHash string,
) (pointAdjustmentPreviewClaims, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	secret := service.pointAdjustmentPreviewSecret()
	if len(parts) != 2 || len(secret) == 0 {
		return pointAdjustmentPreviewClaims{}, appErrors.AdminStateConflict.New("积分调整预览无效或已过期")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return pointAdjustmentPreviewClaims{}, appErrors.AdminStateConflict.New("积分调整预览无效或已过期")
	}
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return pointAdjustmentPreviewClaims{}, appErrors.AdminStateConflict.New("积分调整预览无效或已过期")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return pointAdjustmentPreviewClaims{}, appErrors.AdminStateConflict.New("积分调整预览无效或已过期")
	}
	var claims pointAdjustmentPreviewClaims
	if err := json.Unmarshal(payload, &claims); err != nil ||
		claims.RequestHash != expectedHash ||
		claims.ExpectedUserVersion < 1 ||
		service.now().Unix() > claims.ExpiresAt {
		return pointAdjustmentPreviewClaims{}, appErrors.AdminStateConflict.New("积分调整预览无效或已过期")
	}
	return claims, nil
}

// PreviewPointAdjustment 预览人工增加或扣减积分后的余额。
func (service *PointsService) PreviewPointAdjustment(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	input orderfoodRequest.PointAdjustmentPreviewInput,
) (orderfoodResponse.PointAdjustmentPreview, error) {
	if err := validatePointRuleActor(actor); err != nil {
		return orderfoodResponse.PointAdjustmentPreview{}, err
	}
	input.UserID = strings.TrimSpace(input.UserID)
	input.Reason = strings.TrimSpace(input.Reason)
	if input.UserID == "" ||
		(input.Direction != "credit" && input.Direction != "debit") ||
		input.Amount < 1 || input.Amount > 100000 ||
		len([]rune(input.Reason)) < 4 || len([]rune(input.Reason)) > 200 {
		return orderfoodResponse.PointAdjustmentPreview{}, appErrors.AdminBadRequest.DefaultMsg()
	}
	db := service.database()
	if db == nil {
		return orderfoodResponse.PointAdjustmentPreview{}, appErrors.AdminInternal.DefaultMsg()
	}
	var user orderfoodModel.MiniAppUser
	if err := db.WithContext(ctx).First(&user, "id = ?", input.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orderfoodResponse.PointAdjustmentPreview{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return orderfoodResponse.PointAdjustmentPreview{}, appErrors.AdminInternal.Wrap(err, "load point adjustment user")
	}
	delta := input.Amount
	if input.Direction == "debit" {
		delta = -input.Amount
	}
	balanceAfter := user.Points + delta
	if balanceAfter < 0 {
		return orderfoodResponse.PointAdjustmentPreview{}, appErrors.AdminNegativePoints.DefaultMsg()
	}
	requestHash, err := pointAdjustmentRequestHash(actor.AdministratorID, input)
	if err != nil {
		return orderfoodResponse.PointAdjustmentPreview{}, err
	}
	expiresAt := service.now().Add(pointAdjustmentPreviewTTL)
	token, err := service.signPointAdjustmentPreview(
		requestHash,
		user.Version,
		user.Points,
		expiresAt,
	)
	if err != nil {
		return orderfoodResponse.PointAdjustmentPreview{}, err
	}
	return orderfoodResponse.PointAdjustmentPreview{
		PreviewToken:        token,
		ExpiresAt:           expiresAt,
		User:                subscriptionUserReference(user),
		Direction:           input.Direction,
		Amount:              input.Amount,
		BalanceBefore:       user.Points,
		BalanceAfter:        balanceAfter,
		ExpectedUserVersion: user.Version,
	}, nil
}

// writePointAdjustmentAudit 在余额事务内写入人工积分调整审计。
func (service *PointsService) writePointAdjustmentAudit(
	ctx context.Context,
	tx *gorm.DB,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	reason string,
	userID string,
	before interface{},
	after interface{},
) error {
	if service == nil || service.MutationAudit == nil || tx == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	beforeJSON, err := safeAIAuditJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := safeAIAuditJSON(after)
	if err != nil {
		return err
	}
	key := strings.TrimSpace(idempotencyKey)
	reasonValue := truncateRunes(strings.TrimSpace(reason), 200)
	audit := orderfoodModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                "point_adjustment_create",
		TargetType:            "user_points",
		TargetID:              userID,
		Reason:                &reasonValue,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		RequestID:             truncateRunes(strings.TrimSpace(actor.RequestID), 96),
		IdempotencyKey:        &key,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	return service.MutationAudit.WriteMutationAudit(ctx, tx, &audit)
}

// CreatePointAdjustment 提交人工积分调整并在同一事务写余额、流水和审计。
func (service *PointsService) CreatePointAdjustment(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.PointAdjustmentCreateInput,
) (orderfoodResponse.PointAdjustmentResult, bool, error) {
	if err := validatePointRuleActor(actor); err != nil {
		return orderfoodResponse.PointAdjustmentResult{}, false, err
	}
	previewInput := orderfoodRequest.PointAdjustmentPreviewInput{
		UserID:    strings.TrimSpace(input.UserID),
		Direction: input.Direction,
		Amount:    input.Amount,
		Reason:    strings.TrimSpace(input.Reason),
	}
	requestHash, err := pointAdjustmentRequestHash(actor.AdministratorID, previewInput)
	if err != nil {
		return orderfoodResponse.PointAdjustmentResult{}, false, err
	}
	claims, err := service.verifyPointAdjustmentPreview(input.PreviewToken, requestHash)
	if err != nil {
		return orderfoodResponse.PointAdjustmentResult{}, false, err
	}
	if input.ExpectedUserVersion != claims.ExpectedUserVersion {
		return orderfoodResponse.PointAdjustmentResult{}, false,
			appErrors.AdminStateConflict.New("积分调整预览对应的用户版本已变化，请重新预览")
	}
	input.UserID = previewInput.UserID
	input.Reason = previewInput.Reason
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"point_adjustment_create",
		strings.TrimSpace(idempotencyKey),
		input,
		func(tx *gorm.DB) (interface{}, error) {
			var user orderfoodModel.MiniAppUser
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&user, "id = ?", input.UserID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, appErrors.AdminNotFound.DefaultMsg()
				}
				return nil, appErrors.AdminInternal.Wrap(err, "lock point adjustment user")
			}
			if user.Version != claims.ExpectedUserVersion ||
				user.Points != claims.BalanceBefore {
				return nil, appErrors.AdminStateConflict.New("用户积分或版本已变化，请重新预览")
			}
			delta := input.Amount
			title := "人工增加积分"
			if input.Direction == "debit" {
				delta = -input.Amount
				title = "人工扣减积分"
			}
			balanceBefore := user.Points
			balanceAfter := balanceBefore + delta
			if balanceAfter < 0 {
				return nil, appErrors.AdminNegativePoints.DefaultMsg()
			}
			nextVersion := user.Version + 1
			updateResult := tx.Model(&orderfoodModel.MiniAppUser{}).
				Where("id = ? AND version = ?", user.ID, user.Version).
				Updates(map[string]interface{}{
					"points":     balanceAfter,
					"version":    nextVersion,
					"updated_at": service.now(),
				})
			if updateResult.Error != nil {
				return nil, appErrors.AdminInternal.Wrap(updateResult.Error, "update point adjustment balance")
			}
			if updateResult.RowsAffected != 1 {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			requestID := strings.TrimSpace(actor.RequestID)
			key := strings.TrimSpace(idempotencyKey)
			adminID := actor.AdministratorID
			username := actor.Username
			scene := "manual_adjustment"
			objectType := "user"
			entry := orderfoodModel.FrontPointEntry{
				ID:                    newPublicID("point"),
				UserID:                user.ID,
				Type:                  orderfoodModel.PointAdjustment,
				Scene:                 scene,
				Title:                 title,
				Description:           input.Reason,
				Amount:                delta,
				BalanceAfter:          balanceAfter,
				RelatedObjectType:     &objectType,
				RelatedObjectID:       &user.ID,
				RequestID:             &requestID,
				IdempotencyKey:        &key,
				AdministratorID:       &adminID,
				AdministratorUsername: &username,
				AdministratorNickname: actor.Nickname,
				CreatedAt:             service.now(),
			}
			if err := tx.Create(&entry).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "create point adjustment entry")
			}
			user.Points = balanceAfter
			user.Version = nextVersion
			result := orderfoodResponse.PointAdjustmentResult{
				UserID:        user.ID,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceAfter,
				PointEntry:    pointEntryResponse(entry, user),
				UserVersion:   nextVersion,
			}
			if err := service.writePointAdjustmentAudit(
				ctx, tx, actor, idempotencyKey, input.Reason, user.ID,
				map[string]interface{}{
					"balance": balanceBefore,
					"version": input.ExpectedUserVersion,
				},
				result,
			); err != nil {
				return nil, err
			}
			return result, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.PointAdjustmentResult](raw, replayed, err)
}

// EnsureDefaults 初始化并补齐默认值。
func (service *PointsService) EnsureDefaults(ctx context.Context) error {
	db := service.database()
	if db == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	row := orderfoodModel.PointRuleConfig{
		SingletonKey:       pointRuleSingletonKey,
		Version:            1,
		DailyCheckinReward: DefaultDailyCheckinReward,
		AppliedByID:        0,
		AppliedByUsername:  "system",
		AppliedAt:          service.now(),
		Reason:             "系统默认积分规则",
	}
	if err := db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&row).Error; err != nil {
		return appErrors.AdminInternal.Wrap(err, "seed default point rule")
	}
	return nil
}

func latestPointRule(db *gorm.DB) (orderfoodModel.PointRuleConfig, error) {
	var row orderfoodModel.PointRuleConfig
	if err := db.Where("singleton_key = ?", pointRuleSingletonKey).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orderfoodModel.PointRuleConfig{}, appErrors.AdminNotFound.DefaultMsg()
		}
		return orderfoodModel.PointRuleConfig{}, appErrors.AdminInternal.Wrap(err, "load current point rule")
	}
	return row, nil
}

// CurrentDailyCheckinReward returns the amount to snapshot into a new check-in
// reward transaction. The supplied DB may be an existing transaction.
func CurrentDailyCheckinReward(ctx context.Context, db *gorm.DB) (int64, error) {
	if db == nil {
		db = global.GVA_DB
	}
	if db == nil {
		return 0, appErrors.FrontInternal.DefaultMsg()
	}
	var row orderfoodModel.PointRuleConfig
	if err := db.WithContext(ctx).
		Where("singleton_key = ?", pointRuleSingletonKey).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultDailyCheckinReward, nil
		}
		return 0, appErrors.FrontInternal.Wrap(err, "load daily checkin reward rule")
	}
	if row.DailyCheckinReward < 1 || row.DailyCheckinReward > 100 {
		return 0, appErrors.FrontInternal.DefaultMsg()
	}
	return row.DailyCheckinReward, nil
}

func pointRuleAdministrator(
	id uint,
	username string,
	nickname *string,
) orderfoodResponse.AdministratorSummary {
	return orderfoodResponse.AdministratorSummary{
		ID:       strconv.FormatUint(uint64(id), 10),
		Username: username,
		Nickname: nickname,
	}
}

func pointRuleConfigResponse(row orderfoodModel.PointRuleConfig) orderfoodResponse.PointRuleConfig {
	return orderfoodResponse.PointRuleConfig{
		Version:            row.Version,
		DailyCheckinReward: row.DailyCheckinReward,
		UpdatedBy: pointRuleAdministrator(
			row.AppliedByID,
			row.AppliedByUsername,
			row.AppliedByNickname,
		),
		UpdatedAt: row.AppliedAt,
		Reason:    row.Reason,
	}
}

// CurrentPointRule 获取当前生效的积分规则。
func (service *PointsService) CurrentPointRule(ctx context.Context) (orderfoodResponse.PointRuleConfig, error) {
	if err := service.EnsureDefaults(ctx); err != nil {
		return orderfoodResponse.PointRuleConfig{}, err
	}
	current, err := latestPointRule(service.database().WithContext(ctx))
	if err != nil {
		return orderfoodResponse.PointRuleConfig{}, err
	}
	return pointRuleConfigResponse(current), nil
}

func validatePointRuleActor(actor orderfoodRequest.AdminActor) error {
	if actor.AdministratorID == 0 ||
		strings.TrimSpace(actor.Username) == "" ||
		strings.TrimSpace(actor.RequestID) == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	return nil
}

func (service *PointsService) writeMutationAudit(
	ctx context.Context,
	tx *gorm.DB,
	actor orderfoodRequest.AdminActor,
	action string,
	reason string,
	idempotencyKey string,
	before interface{},
	after interface{},
) error {
	if service == nil || service.MutationAudit == nil || tx == nil {
		return appErrors.AdminInternal.DefaultMsg()
	}
	requestID := truncateRunes(strings.TrimSpace(actor.RequestID), 96)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if requestID == "" || idempotencyKey == "" {
		return appErrors.AdminBadRequest.DefaultMsg()
	}
	beforeJSON, err := safeAIAuditJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := safeAIAuditJSON(after)
	if err != nil {
		return err
	}
	var reasonPointer *string
	if value := truncateRunes(strings.TrimSpace(reason), 200); value != "" {
		reasonPointer = &value
	}
	audit := orderfoodModel.AdminAuditLog{
		AdministratorID:       actor.AdministratorID,
		AdministratorUsername: actor.Username,
		AdministratorNickname: actor.Nickname,
		Action:                action,
		TargetType:            "point_rule",
		TargetID:              pointRuleSingletonKey,
		Reason:                reasonPointer,
		BeforeSummary:         beforeJSON,
		AfterSummary:          afterJSON,
		RequestID:             requestID,
		IdempotencyKey:        &idempotencyKey,
	}
	if value := strings.TrimSpace(actor.SourceIPMasked); value != "" {
		audit.SourceIPMasked = &value
	}
	if value := truncateRunes(strings.TrimSpace(actor.UserAgentSummary), 240); value != "" {
		audit.UserAgentSummary = &value
	}
	return service.MutationAudit.WriteMutationAudit(ctx, tx, &audit)
}

// UpdatePointRule 保存并立即应用积分规则。
func (service *PointsService) UpdatePointRule(
	ctx context.Context,
	actor orderfoodRequest.AdminActor,
	idempotencyKey string,
	input orderfoodRequest.PointRuleUpdateInput,
) (orderfoodResponse.PointRuleConfig, bool, error) {
	if err := validatePointRuleActor(actor); err != nil {
		return orderfoodResponse.PointRuleConfig{}, false, err
	}
	reason := strings.TrimSpace(input.Reason)
	if input.DailyCheckinReward < 1 || input.DailyCheckinReward > 100 ||
		input.ExpectedVersion < 1 || len([]rune(reason)) < 2 || len([]rune(reason)) > 200 {
		return orderfoodResponse.PointRuleConfig{}, false, appErrors.AdminBadRequest.DefaultMsg()
	}
	if err := service.EnsureDefaults(ctx); err != nil {
		return orderfoodResponse.PointRuleConfig{}, false, err
	}
	raw, replayed, err := service.Idempotency.Execute(
		ctx,
		actor.AdministratorID,
		"point_rule_update",
		idempotencyKey,
		input,
		func(tx *gorm.DB) (interface{}, error) {
			current, err := latestPointRule(tx.Clauses(clause.Locking{Strength: "UPDATE"}))
			if err != nil {
				return nil, err
			}
			if current.Version != input.ExpectedVersion {
				return nil, appErrors.AdminStateConflict.DefaultMsg()
			}
			next := orderfoodModel.PointRuleConfig{
				SingletonKey:       pointRuleSingletonKey,
				Version:            current.Version + 1,
				DailyCheckinReward: input.DailyCheckinReward,
				AppliedByID:        actor.AdministratorID,
				AppliedByUsername:  actor.Username,
				AppliedByNickname:  actor.Nickname,
				AppliedAt:          service.now(),
				Reason:             reason,
			}
			if err := tx.Save(&next).Error; err != nil {
				return nil, appErrors.AdminInternal.Wrap(err, "save point rule")
			}
			before := pointRuleConfigResponse(current)
			after := pointRuleConfigResponse(next)
			if err := service.writeMutationAudit(
				ctx,
				tx,
				actor,
				"update_point_rule",
				reason,
				idempotencyKey,
				before,
				after,
			); err != nil {
				return nil, err
			}
			return after, nil
		},
	)
	return decodeIdempotentResult[orderfoodResponse.PointRuleConfig](raw, replayed, err)
}
