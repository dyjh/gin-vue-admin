package orderfood

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"github.com/dyjh/order-food-mini-app/server/global"
	"github.com/dyjh/order-food-mini-app/server/middleware"
	orderfoodModel "github.com/dyjh/order-food-mini-app/server/model/orderfood"
	systemModel "github.com/dyjh/order-food-mini-app/server/model/system"
	systemRequest "github.com/dyjh/order-food-mini-app/server/model/system/request"
	orderfoodService "github.com/dyjh/order-food-mini-app/server/service/orderfood"
	"github.com/dyjh/order-food-mini-app/server/testutil"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type apiFakeAIConnector struct{}

func (apiFakeAIConnector) TestConnection(
	context.Context,
	orderfoodModel.AIProvider,
	time.Duration,
) (orderfoodService.AIConnectionResult, error) {
	return orderfoodService.AIConnectionResult{
		Success:     true,
		Category:    "ok",
		SafeMessage: "连接成功",
	}, nil
}

func (apiFakeAIConnector) TestPrompt(
	context.Context,
	orderfoodModel.AIProvider,
	orderfoodModel.AIModel,
	orderfoodService.AIPromptRunRequest,
	time.Duration,
) (orderfoodService.AIPromptRunResult, error) {
	return orderfoodService.AIPromptRunResult{
		Success:     true,
		SafeMessage: "测试成功",
	}, nil
}

// openAIApiTestDB 创建 AI 管理接口测试所需的独立 MySQL 数据库。
func openAIApiTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	models := append(
		orderfoodService.AIModelsForMigration(),
		&orderfoodModel.AdminIdempotencyRecord{},
		&orderfoodModel.AdminAccessAudit{},
		&orderfoodModel.AdminAuditLog{},
		&orderfoodModel.MiniAppUser{},
		&systemModel.SysBaseMenuBtn{},
		&systemModel.SysAuthorityBtn{},
	)
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate AI API models: %v", err)
	}
	for authorityID, permissions := range orderfoodService.DefaultRolePermissionMatrix() {
		for _, permission := range permissions {
			grantAIApiButtonPermission(t, db, authorityID, permission)
		}
	}
	return db
}

// grantAIApiButtonPermission 为接口测试角色注入一项 GVA 按钮权限。
func grantAIApiButtonPermission(
	t *testing.T,
	db *gorm.DB,
	authorityID uint,
	permission string,
) {
	t.Helper()
	var button systemModel.SysBaseMenuBtn
	if err := db.Where("name = ?", permission).FirstOrCreate(
		&button,
		systemModel.SysBaseMenuBtn{
			Name:          permission,
			Desc:          permission,
			SysBaseMenuID: 1,
		},
	).Error; err != nil {
		t.Fatalf("create API test button permission %q: %v", permission, err)
	}
	authorization := systemModel.SysAuthorityBtn{
		AuthorityId:      authorityID,
		SysMenuID:        button.SysBaseMenuID,
		SysBaseMenuBtnID: button.ID,
	}
	if err := db.Where(
		"authority_id = ? AND sys_menu_id = ? AND sys_base_menu_btn_id = ?",
		authorization.AuthorityId,
		authorization.SysMenuID,
		authorization.SysBaseMenuBtnID,
	).FirstOrCreate(&authorization).Error; err != nil {
		t.Fatalf("grant API test button permission %q: %v", permission, err)
	}
}

func aiAPIEngine(t *testing.T, service *orderfoodService.AIService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(func(c *gin.Context) {
		authorityID := orderfoodModel.AuthorityOrderFoodSuperAdmin
		if raw := c.GetHeader("X-Test-Authority"); raw != "" {
			parsed, err := strconv.ParseUint(raw, 10, 64)
			if err == nil {
				authorityID = uint(parsed)
			}
		}
		c.Set("claims", &systemRequest.CustomClaims{
			BaseClaims: systemRequest.BaseClaims{
				ID:          9,
				Username:    "api-admin",
				NickName:    "API 管理员",
				AuthorityId: authorityID,
			},
		})
		c.Next()
	})
	api := NewAIApi(service)
	group := engine.Group("/orderfood")
	group.POST("/ai-providers", api.CreateAIProvider)
	group.GET("/ai-providers/:providerId", api.GetAIProvider)
	group.GET("/ai-capabilities/:capabilityCode", api.GetAICapability)
	group.GET("/ai-capabilities/:capabilityCode/prompts", api.GetAIPromptWorkspace)
	return engine
}

type aiAPIEnvelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
	Msg  string          `json:"msg"`
}

func performAIAPIRequest(
	t *testing.T,
	engine http.Handler,
	method string,
	path string,
	body string,
	headers map[string]string,
) (*httptest.ResponseRecorder, aiAPIEnvelope) {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	var envelope aiAPIEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response %s %s: %v; body=%s", method, path, err, recorder.Body.String())
	}
	return recorder, envelope
}

func TestAIApiCredentialIsWriteOnlyAndPromptBodiesArePermissioned(t *testing.T) {
	previousValidator := global.GVA_VALIDATOR
	global.GVA_VALIDATOR = validator.New()
	defer func() {
		global.GVA_VALIDATOR = previousValidator
	}()

	db := openAIApiTestDB(t)
	service := orderfoodService.NewAIService(
		db,
		orderfoodService.NewPermissionService(db),
		orderfoodService.NewIdempotencyService(db),
		apiFakeAIConnector{},
	)
	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure AI defaults: %v", err)
	}
	grantAIApiButtonPermission(
		t,
		db,
		orderfoodModel.AuthorityOrderFoodOperator,
		orderfoodService.PermissionAICapabilityRead,
	)
	engine := aiAPIEngine(t, service)

	createBody := `{
		"name":"DeepSeek",
		"type":"deepseek",
		"baseUrl":"https://api.deepseek.example/v1",
		"credentialRef":"env://DEEPSEEK_API_KEY",
		"timeoutMs":30000,
		"retryCount":1,
		"enabled":true
	}`
	recorder, envelope := performAIAPIRequest(
		t,
		engine,
		http.MethodPost,
		"/orderfood/ai-providers",
		createBody,
		map[string]string{
			"X-Idempotency-Key": "api-provider-create",
			"X-Request-Id":      "request-provider-create",
		},
	)
	if recorder.Code != http.StatusOK || envelope.Code != 0 {
		t.Fatalf("create provider response: status=%d envelope=%+v", recorder.Code, envelope)
	}
	if strings.Contains(recorder.Body.String(), "DEEPSEEK_API_KEY") ||
		strings.Contains(recorder.Body.String(), "credentialRef") {
		t.Fatalf("API response leaked credential reference: %s", recorder.Body.String())
	}
	var provider struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(envelope.Data, &provider); err != nil || provider.ID == "" {
		t.Fatalf("decode provider ID: %+v err=%v", provider, err)
	}
	_, detailEnvelope := performAIAPIRequest(
		t,
		engine,
		http.MethodGet,
		"/orderfood/ai-providers/"+provider.ID,
		"",
		map[string]string{"X-Request-Id": "request-provider-detail"},
	)
	if detailEnvelope.Code != 0 ||
		strings.Contains(string(detailEnvelope.Data), "DEEPSEEK_API_KEY") ||
		strings.Contains(string(detailEnvelope.Data), "credentialRef") {
		t.Fatalf("provider detail leaked credential: %+v", detailEnvelope)
	}

	_, missingKeyEnvelope := performAIAPIRequest(
		t,
		engine,
		http.MethodPost,
		"/orderfood/ai-providers",
		createBody,
		map[string]string{"X-Request-Id": "request-missing-key"},
	)
	if missingKeyEnvelope.Code != int(appErrors.AdminBadRequest) {
		t.Fatalf("missing idempotency key code = %d", missingKeyEnvelope.Code)
	}

	operatorHeaders := map[string]string{
		"X-Test-Authority": strconv.FormatUint(
			uint64(orderfoodModel.AuthorityOrderFoodOperator),
			10,
		),
		"X-Request-Id": "request-capability-operator",
	}
	_, redactedEnvelope := performAIAPIRequest(
		t,
		engine,
		http.MethodGet,
		"/orderfood/ai-capabilities/meal_suggest",
		"",
		operatorHeaders,
	)
	if redactedEnvelope.Code != 0 ||
		strings.Contains(string(redactedEnvelope.Data), "systemPrompt") ||
		strings.Contains(string(redactedEnvelope.Data), "userPromptTemplate") {
		t.Fatalf("operator capability response exposed prompt bodies: %+v", redactedEnvelope)
	}
	_, deniedPromptEnvelope := performAIAPIRequest(
		t,
		engine,
		http.MethodGet,
		"/orderfood/ai-capabilities/meal_suggest/prompts",
		"",
		operatorHeaders,
	)
	if deniedPromptEnvelope.Code != int(appErrors.AdminNoPermission) {
		t.Fatalf("operator prompt workspace code = %d", deniedPromptEnvelope.Code)
	}
	_, fullPromptEnvelope := performAIAPIRequest(
		t,
		engine,
		http.MethodGet,
		"/orderfood/ai-capabilities/meal_suggest/prompts",
		"",
		map[string]string{"X-Request-Id": "request-capability-super"},
	)
	if fullPromptEnvelope.Code != 0 ||
		!strings.Contains(string(fullPromptEnvelope.Data), "defaultSystemPrompt") ||
		!strings.Contains(string(fullPromptEnvelope.Data), "defaultUserPromptTemplate") {
		t.Fatalf("super prompt workspace incomplete: %+v", fullPromptEnvelope)
	}
	var accessCount int64
	if err := db.Model(&orderfoodModel.AdminAccessAudit{}).Count(&accessCount).Error; err != nil {
		t.Fatalf("count prompt access audits: %v", err)
	}
	if accessCount != 1 {
		t.Fatalf("prompt access audit count = %d, want 1", accessCount)
	}
}
