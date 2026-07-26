package errors

import (
	stderrors "errors"
	"fmt"

	pkgerrors "github.com/pkg/errors"
)

// ErrorType is the stable business error code returned in the response envelope.
// Admin codes are kept in 20000-29999 and mini-program codes in 40000-49999.
type ErrorType int

const (
	SUCCESS ErrorType = 0

	AdminBadRequest          ErrorType = 20001
	AdminLoginExpired        ErrorType = 20101
	AdminNoPermission        ErrorType = 20301
	AdminAccountUnavailable  ErrorType = 20302
	AdminNotFound            ErrorType = 20401
	AdminStateConflict       ErrorType = 20901
	AdminIdempotencyConflict ErrorType = 20902
	AdminAlreadyExists       ErrorType = 20903
	AdminRequestProcessing   ErrorType = 20904
	AdminResourceInUse       ErrorType = 20905
	AdminImageRejected       ErrorType = 22201
	AdminInvalidImage        ErrorType = 22202
	AdminNegativePoints      ErrorType = 22203
	AdminInvalidConfig       ErrorType = 22204
	AdminInvalidGovernance   ErrorType = 22205
	AdminIdentityKeyMissing  ErrorType = 22206
	AdminIdentityKeyInvalid  ErrorType = 22207
	AdminRateLimited         ErrorType = 22901
	AdminInternal            ErrorType = 25000
	AdminProviderFailed      ErrorType = 25001
	AdminProviderTimeout     ErrorType = 25004

	FrontBadRequest          ErrorType = 40001
	FrontLoginExpired        ErrorType = 40101
	FrontWxLoginInvalid      ErrorType = 40102
	FrontNoPermission        ErrorType = 40301
	FrontUserDisabled        ErrorType = 40302
	FrontNotFound            ErrorType = 40401
	FrontStateConflict       ErrorType = 40901
	FrontIdempotencyConflict ErrorType = 40902
	FrontAlreadyExists       ErrorType = 40903
	FrontRequestProcessing   ErrorType = 40904
	FrontImageRejected       ErrorType = 42201
	FrontInvalidImage        ErrorType = 42202
	FrontRateLimited         ErrorType = 42901
	FrontInternal            ErrorType = 45000
	FrontFeatureDisabled     ErrorType = 46001
	FrontPointsDisabled      ErrorType = 46002
	FrontInsufficientPoints  ErrorType = 47001
	FrontFeatureLocked       ErrorType = 47002
	FrontExecutionRefunded   ErrorType = 47003
	FrontQuotaExceeded       ErrorType = 47004
	FrontResultInvalid       ErrorType = 47005
	FrontRefundPending       ErrorType = 47006

	// Backward-compatible aliases for the custom error API introduced earlier.
	NoType       = AdminInternal
	BadRequest   = AdminBadRequest
	DataNotFound = AdminNotFound
	ValidateFail = AdminBadRequest
	NoAuth       = AdminLoginExpired
	NoPermission = AdminNoPermission
	DefaultError = AdminInternal
)

var errorMessages = map[ErrorType]string{
	AdminBadRequest:          "请求参数错误",
	AdminLoginExpired:        "管理员登录已过期",
	AdminNoPermission:        "缺少 API 或按钮权限",
	AdminAccountUnavailable:  "管理员账号不可用",
	AdminNotFound:            "资源不存在或不可见",
	AdminStateConflict:       "目标状态或版本冲突",
	AdminIdempotencyConflict: "幂等键对应的请求参数不同",
	AdminAlreadyExists:       "资源已生成、已发布或已存在",
	AdminRequestProcessing:   "相同幂等请求仍在处理中",
	AdminResourceInUse:       "资源正在被引用，不能删除或停用",
	AdminImageRejected:       "图片内容审核未通过",
	AdminInvalidImage:        "图片资源无效、未审核通过或用途不匹配",
	AdminNegativePoints:      "积分调整后余额小于 0",
	AdminInvalidConfig:       "配置组合无效或缺少可用依赖",
	AdminInvalidGovernance:   "违规处理动作或影响范围不合法",
	AdminIdentityKeyMissing:  "服务端未配置数据加密密钥，请在 config.yaml 的 orderfood.identity-key 中配置并重启服务",
	AdminIdentityKeyInvalid:  "服务端数据加密密钥格式错误，orderfood.identity-key 必须是32字节密钥的Base64编码",
	AdminRateLimited:         "操作过于频繁",
	AdminInternal:            "未预期服务器错误",
	AdminProviderFailed:      "外部供应商调用失败",
	AdminProviderTimeout:     "外部供应商调用超时",

	FrontBadRequest:          "请求参数错误",
	FrontLoginExpired:        "登录已过期",
	FrontWxLoginInvalid:      "微信登录凭证无效、已使用或兑换失败",
	FrontNoPermission:        "无权操作该资源",
	FrontUserDisabled:        "用户已被禁用",
	FrontNotFound:            "资源不存在",
	FrontStateConflict:       "状态冲突或重复操作",
	FrontIdempotencyConflict: "幂等请求参数与原请求不一致",
	FrontAlreadyExists:       "资源已复制、已生成或已存在",
	FrontRequestProcessing:   "相同幂等请求仍在处理中",
	FrontImageRejected:       "图片内容审核不通过",
	FrontInvalidImage:        "图片资源无效、未审核通过或不属于当前用户",
	FrontRateLimited:         "请求过于频繁",
	FrontInternal:            "未预期服务器错误",
	FrontFeatureDisabled:     "当前功能未开放",
	FrontPointsDisabled:      "积分功能未开放",
	FrontInsufficientPoints:  "积分不足",
	FrontFeatureLocked:       "当前功能未解锁",
	FrontExecutionRefunded:   "处理失败，积分已退还",
	FrontQuotaExceeded:       "免费次数或日限额已用完",
	FrontResultInvalid:       "结果业务校验失败",
	FrontRefundPending:       "处理失败，积分退还处理中",
}

// ErrorContext is safe structured context. Do not put credentials, tokens,
// WeChat identifiers, prompts, provider payloads or private content here.
type ErrorContext struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type customError struct {
	errorType     ErrorType
	originalError error
	context       []ErrorContext
}

func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

func (errorType ErrorType) DefaultMsg() error {
	return customError{errorType: errorType, originalError: stderrors.New(Message(errorType))}
}

func (errorType ErrorType) New(msg string) error {
	if msg == "" {
		msg = Message(errorType)
	}
	return customError{errorType: errorType, originalError: stderrors.New(msg)}
}

func (errorType ErrorType) Append(msg string) error {
	base := Message(errorType)
	if msg == "" {
		return errorType.DefaultMsg()
	}
	return customError{errorType: errorType, originalError: fmt.Errorf("%s: %s", base, msg)}
}

func (errorType ErrorType) NewTrace(msg string) error {
	if msg == "" {
		msg = Message(errorType)
	}
	return customError{errorType: errorType, originalError: pkgerrors.WithStack(stderrors.New(msg))}
}

func (errorType ErrorType) Newf(msg string, args ...interface{}) error {
	return customError{errorType: errorType, originalError: fmt.Errorf(msg, args...)}
}

func (errorType ErrorType) Wrap(err error, msg string) error {
	return errorType.Wrapf(err, "%s", msg)
}

func (errorType ErrorType) Wrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return customError{
		errorType:     errorType,
		originalError: pkgerrors.Wrapf(err, msg, args...),
	}
}

func (error customError) Error() string {
	if error.originalError == nil {
		return Message(error.errorType)
	}
	return error.originalError.Error()
}

func (error customError) Unwrap() error {
	return error.originalError
}

func (error customError) Code() int {
	return int(error.errorType)
}

func (error customError) Type() ErrorType {
	return error.errorType
}

func (error customError) Context() []ErrorContext {
	return append([]ErrorContext(nil), error.context...)
}

func (error customError) AddErrorContext(field, message string) error {
	error.context = append(error.context, ErrorContext{Field: field, Message: message})
	return error
}

func (error customError) Format(s fmt.State, verb rune) {
	if formatter, ok := error.originalError.(fmt.Formatter); ok {
		formatter.Format(s, verb)
		return
	}
	fmt.Fprint(s, error.Error())
}

// New creates an internal error while retaining the previous package API.
func New(msg string) error {
	return AdminInternal.New(msg)
}

func Newf(msg string, args ...interface{}) error {
	return AdminInternal.Newf(msg, args...)
}

func NewCode(code ErrorType, msg string) error {
	return code.New(msg)
}

func Wrap(err error, msg string) error {
	return Wrapf(err, "%s", msg)
}

func Wrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	var typed customError
	if stderrors.As(err, &typed) {
		return customError{
			errorType:     typed.errorType,
			originalError: pkgerrors.Wrapf(err, msg, args...),
			context:       typed.Context(),
		}
	}
	return AdminInternal.Wrapf(err, msg, args...)
}

func Cause(err error) error {
	return pkgerrors.Cause(err)
}

func AddErrorContext(err error, field, message string) error {
	if err == nil {
		return nil
	}
	var typed customError
	if stderrors.As(err, &typed) {
		return typed.AddErrorContext(field, message)
	}
	return customError{
		errorType:     AdminInternal,
		originalError: err,
		context:       []ErrorContext{{Field: field, Message: message}},
	}
}

func GetErrorContext(err error) []ErrorContext {
	var typed customError
	if stderrors.As(err, &typed) {
		return typed.Context()
	}
	return nil
}

func GetType(err error) ErrorType {
	var typed customError
	if stderrors.As(err, &typed) {
		return typed.errorType
	}
	return AdminInternal
}

func Code(err error) int {
	return int(GetType(err))
}

func Message(code ErrorType) string {
	if message, ok := errorMessages[code]; ok {
		return message
	}
	return "未预期服务器错误"
}

// SafeMessage returns the stable message registered for coded errors. It never
// exposes wrapped provider, SQL or infrastructure error text to the client.
func SafeMessage(err error) string {
	return Message(GetType(err))
}
