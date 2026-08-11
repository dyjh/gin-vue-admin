package ai

import (
	"time"
)

// AIModelCapability 表示能力定义要求模型支持的输入输出模态。
type AIModelCapability string

const (
	AIModelText            AIModelCapability = "text"
	AIModelVision          AIModelCapability = "vision"
	AIModelImageGeneration AIModelCapability = "image_generation"
)

const (
	AICapabilityDishTextExtract     = "dish_text_extract"
	AICapabilityRecipeImageExtract  = "recipe_image_extract"
	AICapabilityDishCoverCreate     = "dish_cover_create"
	AICapabilityCheckinImageAnalyze = "checkin_image_analyze"
	AICapabilityPreferenceSummarize = "preference_profile_summarize"
	AICapabilityMealSuggest         = "meal_suggest"
	AICapabilityPrepSequence        = "prep_sequence"
)

// AICapabilityDefinition contains stable product metadata. Client-facing
// labels remain in the platform policy and are not duplicated here.
type AICapabilityDefinition struct {
	Code                             string             `json:"code" gorm:"column:code;type:varchar(64);primaryKey;not null;comment:AI能力编码;"`                                                       // AI能力编码
	Name                             string             `json:"name" gorm:"column:name;type:varchar(80);not null;comment:名称;"`                                                                      // 名称
	ClientFeatureCode                string             `json:"clientFeatureCode" gorm:"column:client_feature_code;type:varchar(40);not null;index;comment:小程序功能编码;"`                               // 小程序功能编码
	RequiredModelCapability          AIModelCapability  `json:"requiredModelCapability" gorm:"column:required_model_capability;type:varchar(32);not null;comment:主模型所需能力;"`                         // 主模型所需能力
	RequiredAuxiliaryModelCapability *AIModelCapability `json:"requiredAuxiliaryModelCapability" gorm:"column:required_auxiliary_model_capability;type:varchar(32);default:null;comment:辅助模型所需能力;"` // 辅助模型所需能力
	SortOrder                        int                `json:"sortOrder" gorm:"column:sort_order;type:int;not null;index;comment:排序值;"`                                                            // 排序值
	CreatedAt                        time.Time          `json:"-" gorm:"column:created_at;type:datetime;not null;comment:创建时间;"`                                                                    // 创建时间
	UpdatedAt                        time.Time          `json:"updatedAt" gorm:"column:updated_at;type:datetime;not null;comment:更新时间;"`                                                            // 更新时间
}

// TableName 指定AICapabilityDefinition对应的数据表名。
func (AICapabilityDefinition) TableName() string {
	return "of_ai_capabilities"
}
