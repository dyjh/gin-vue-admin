package experience

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	userModel "github.com/dyjh/order-food-mini-app/server/model/user"
	userResponse "github.com/dyjh/order-food-mini-app/server/model/user/response"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"gorm.io/gorm"
)

// openPreferenceTestDB 创建偏好画像测试所需的独立 MySQL 数据库。
func openPreferenceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(
		&userModel.UserPreferenceProfile{},
		&userModel.PreferenceEvidenceAggregate{},
		&userModel.PreferenceEvidence{},
	); err != nil {
		t.Fatalf("migrate preference schema: %v", err)
	}
	return db
}

// TestPreferencePipelineCoversSevenSourcesAndKeepsPendingState 验证七类证据均可入队且未完成图片分析时画像保持更新中。
func TestPreferencePipelineCoversSevenSourcesAndKeepsPendingState(t *testing.T) {
	db := openPreferenceTestDB(t)
	now := time.Date(2026, 7, 26, 9, 0, 0, 0, time.UTC)
	service := &PreferenceService{DB: db, Now: func() time.Time { return now }}
	ctx := context.Background()
	userID := "preference-user"
	type evidenceCase struct {
		sourceType string          // 证据来源
		sourceID   string          // 来源业务ID
		direction  string          // 正向或负向
		weight     float64         // 聚合权重
		facts      preferenceFacts // 结构化事实
	}
	cases := []evidenceCase{
		{
			sourceType: userModel.PreferenceSourceCheckinImage,
			sourceID:   "checkin-1",
			direction:  "positive",
			weight:     1.5,
			facts: preferenceFacts{
				ImageURL: "/uploads/checkin.png",
				Caption:  "番茄炒蛋",
			},
		},
		{
			sourceType: userModel.PreferenceSourceRecommendationAdopted,
			sourceID:   "recommendation-1",
			direction:  "positive",
			weight:     3,
			facts:      preferenceFacts{Tags: []string{"家常菜"}, Ingredients: []string{"番茄"}},
		},
		{
			sourceType: userModel.PreferenceSourceDishSaved,
			sourceID:   "dish-1",
			direction:  "positive",
			weight:     2.5,
			facts:      preferenceFacts{Tags: []string{"快手菜"}, Ingredients: []string{"鸡蛋"}},
		},
		{
			sourceType: userModel.PreferenceSourceDishOrdered,
			sourceID:   "meal-1:candidate-1",
			direction:  "positive",
			weight:     2.5,
			facts:      preferenceFacts{Tags: []string{"川菜"}},
		},
		{
			sourceType: userModel.PreferenceSourceReshuffle,
			sourceID:   "suggestion-1",
			direction:  "negative",
			weight:     0.75,
			facts:      preferenceFacts{Tags: []string{"香辣"}},
		},
		{
			sourceType: userModel.PreferenceSourceSkipForNow,
			sourceID:   "suggestion-2",
			direction:  "negative",
			weight:     1.5,
			facts:      preferenceFacts{Tags: []string{"油炸"}},
		},
		{
			sourceType: userModel.PreferenceSourceExplicitSetting,
			sourceID:   "setting-1",
			direction:  "positive",
			weight:     4,
			facts: preferenceFacts{Settings: []preferenceSetting{{
				Type: "avoidance", Value: "不吃花生", UpdatedAt: now,
			}}},
		},
	}
	for _, item := range cases {
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return service.queueEvidenceTx(
				ctx,
				tx,
				userID,
				item.sourceType,
				item.sourceID,
				item.direction,
				item.weight,
				1,
				item.facts,
				now,
			)
		})
		if err != nil {
			t.Fatalf("queue %s evidence: %v", item.sourceType, err)
		}
	}

	var aggregateCount int64
	if err := db.Model(&userModel.PreferenceEvidenceAggregate{}).
		Where("user_id = ?", userID).
		Count(&aggregateCount).Error; err != nil {
		t.Fatalf("count preference aggregates: %v", err)
	}
	if aggregateCount != int64(len(cases)) {
		t.Fatalf("preference aggregate source count = %d, want %d", aggregateCount, len(cases))
	}

	var rows []userModel.PreferenceEvidence
	if err := db.Where("user_id = ? AND source_type <> ?",
		userID, userModel.PreferenceSourceCheckinImage).
		Order("source_type asc").
		Find(&rows).Error; err != nil {
		t.Fatalf("load rule-based preference evidence: %v", err)
	}
	for _, row := range rows {
		if err := service.processPreferenceEvidence(ctx, row); err != nil {
			t.Fatalf("process %s evidence: %v", row.SourceType, err)
		}
	}

	var profile userModel.UserPreferenceProfile
	if err := db.First(&profile, "user_id = ?", userID).Error; err != nil {
		t.Fatalf("load preference profile: %v", err)
	}
	if !profile.HasProfile || profile.UpdateState != userModel.PreferenceUpdatePending {
		t.Fatalf("preference profile state = %+v", profile)
	}
	var content userResponse.PreferenceProfileContent
	if err := json.Unmarshal(profile.ProfileJSON, &content); err != nil {
		t.Fatalf("decode preference profile content: %v", err)
	}
	if len(content.UserSettings) != 1 ||
		content.UserSettings[0].Type != "avoidance" ||
		content.UserSettings[0].Value != "不吃花生" {
		t.Fatalf("preference user settings = %+v", content.UserSettings)
	}

	var pendingCheckin userModel.PreferenceEvidence
	if err := db.First(&pendingCheckin,
		"user_id = ? AND source_type = ?", userID, userModel.PreferenceSourceCheckinImage).Error; err != nil {
		t.Fatalf("load pending checkin evidence: %v", err)
	}
	if pendingCheckin.Status != userModel.PreferenceEvidencePending {
		t.Fatalf("checkin evidence status = %q, want %q",
			pendingCheckin.Status, userModel.PreferenceEvidencePending)
	}
}
