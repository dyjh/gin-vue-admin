package errors

import (
	stderrors "errors"
	"net/http"
	"strings"
	"testing"
)

func TestCodeRangesAndWrapping(t *testing.T) {
	adminErr := AdminNotFound.Wrap(stderrors.New("sql detail"), "load user")
	if got := Code(adminErr); got != 20401 {
		t.Fatalf("admin code = %d, want 20401", got)
	}
	if got := SafeMessage(adminErr); got != "资源不存在或不可见" {
		t.Fatalf("safe admin message = %q", got)
	}

	frontErr := FrontUserDisabled.DefaultMsg()
	if got := Code(frontErr); got != 40302 {
		t.Fatalf("front code = %d, want 40302", got)
	}
	if got := SafeMessage(frontErr); got != "用户已被禁用" {
		t.Fatalf("safe front message = %q", got)
	}
}

func TestErrorContextSurvivesWrapping(t *testing.T) {
	err := AddErrorContext(AdminBadRequest.DefaultMsg(), "status", "非法状态")
	err = Wrap(err, "binding failed")
	contextItems := GetErrorContext(err)
	if len(contextItems) != 1 || contextItems[0].Field != "status" {
		t.Fatalf("unexpected context: %#v", contextItems)
	}
}

// TestIdentityKeyErrorsHaveActionableMessages 验证管理端能区分密钥缺失和格式错误。
func TestIdentityKeyErrorsHaveActionableMessages(t *testing.T) {
	tests := []struct {
		errorType ErrorType // 待验证的错误类型
		code      int       // 预期业务错误码
		keyword   string    // 提示中必须包含的操作指引
	}{
		{errorType: AdminIdentityKeyMissing, code: 22206, keyword: "config.yaml"},
		{errorType: AdminIdentityKeyInvalid, code: 22207, keyword: "Base64"},
	}
	for _, test := range tests {
		err := test.errorType.DefaultMsg()
		if Code(err) != test.code {
			t.Fatalf("code = %d, want %d", Code(err), test.code)
		}
		if !strings.Contains(SafeMessage(err), test.keyword) {
			t.Fatalf("message = %q, want keyword %q", SafeMessage(err), test.keyword)
		}
		if httpStatusFor(test.errorType) != http.StatusUnprocessableEntity {
			t.Fatalf("HTTP status = %d, want %d", httpStatusFor(test.errorType), http.StatusUnprocessableEntity)
		}
	}
}
