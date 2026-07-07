package utils

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/dyjh/order-food-mini-app/server/global"
	"github.com/go-playground/validator/v10"
)

// VerifyAll 验证已经绑定的查询参数
func VerifyAll(st interface{}) error {
	if global.GVA_VALIDATOR == nil {
		return errors.New("validator not initialized")
	}

	// 假设绑定操作在外部完成，这里只进行验证
	if err := global.GVA_VALIDATOR.Struct(st); err != nil {
		var validationErrors validator.ValidationErrors
		if !errors.As(err, &validationErrors) {
			return err
		}
		for _, errItem := range validationErrors {
			if global.GVA_TRANS != nil {
				return errors.New(errItem.Translate(global.GVA_TRANS))
			}
			return errItem
		}
	}

	if err := CheckSQLInjection(st); err != nil {
		return err
	}

	return nil
}

// IsEmail 判断是否为邮箱
func IsEmail(s string) bool {
	// 邮箱的正则表达式
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(s)
}

// IsPhoneNumber 判断是否为电话号码
func IsPhoneNumber(s string) bool {
	// 电话号码的正则表达式（示例为中国大陆手机号格式）
	var phoneRegex = regexp.MustCompile(`^1\d{10}$`)
	return phoneRegex.MatchString(s)
}

// required: 字段必须存在且不能为空
// required_if: 当另一个字段满足特定条件时，该字段为必填
// required_unless: 除非另一个字段满足特定条件，否则该字段为必填
// required_with: 当指定的字段存在时，该字段为必填
// required_with_all: 当所有指定的字段存在时，该字段为必填
// required_without: 当指定的字段不存在时，该字段为必填
// required_without_all: 当所有指定的字段不存在时，该字段为必填

// len=<number>: 字段长度必须等于指定的数值
// min=<number>: 字段长度或数值必须大于等于指定的数值
// max=<number>: 字段长度或数值必须小于等于指定的数值
// gt=<number>: 字段值必须大于指定的数值
// gte=<number>: 字段值必须大于或等于指定的数值
// lt=<number>: 字段值必须小于指定的数值
// lte=<number>: 字段值必须小于或等于指定的数值

// email: 必须是有效的电子邮件格式
// phone: 必须是有效的手机号码
// url: 必须是有效的 URL 格式
// uri: 必须是有效的 URI 格式
// uuid: 必须是有效的 UUID 格式
// uuid3: 必须是有效的 UUIDv3 格式
// uuid4: 必须是有效的 UUIDv4 格式
// uuid5: 必须是有效的 UUIDv5 格式
// ip: 必须是有效的 IP 地址（支持 IPv4 和 IPv6）
// ipv4: 必须是有效的 IPv4 地址
// ipv6: 必须是有效的 IPv6 地址
// alpha: 必须只包含字母
// alphanum: 必须只包含字母和数字
// numeric: 必须只包含数字
// hexadecimal: 必须是有效的十六进制数
// base64: 必须是有效的 Base64 编码
// ascii: 必须是 ASCII 字符
// printable: 必须是可打印的字符
// multibyte: 必须包含多个字节的字符
// datauri: 必须是有效的数据 URI
// contains=<substr>: 必须包含指定的子字符串
// containsany=<chars>: 必须包含指定的任意字符
// containsrune=<rune>: 必须包含指定的 Unicode 字符
// excludes=<chars>: 必须不包含指定的字符
// excludesall=<chars>: 必须不包含指定的任何字符
// excludesrune=<rune>: 必须不包含指定的 Unicode 字符

// datetime=<format>: 必须符合指定的时间格式。例如，datetime=2006-01-02
// oneof=<option1> <option2> ...: 字段值必须是指定选项中的一个
// notoneof=<option1> <option2> ...: 字段值不能是指定选项中的任何一个
// eqfield=<field>: 字段值必须等于指定字段的值
// eqcsfield=<field>: 字段值必须等于指定字段的值（区分大小写）
// neqfield=<field>: 字段值不能等于指定字段的值
// neqcsfield=<field>: 字段值不能等于指定字段的值（区分大小写）

// unique: 切片中的所有元素必须唯一
// unique=<param>: 根据指定参数检查唯一性
// unique_slices: 切片中的子切片必须唯一

// omitempty: 如果字段为空，则跳过验证
// required_if: 当另一个字段满足特定条件时，该字段为必填
// required_unless: 除非另一个字段满足特定条件，否则该字段为必填
// required_with: 当指定的字段存在时，该字段为必填
// required_with_all: 当所有指定的字段存在时，该字段为必填
// required_without: 当指定的字段不存在时，该字段为必填
// required_without_all: 当所有指定的字段不存在时，该字段为必填

// dive: 用于嵌套结构或数组/切片中的每个元素进行验证
// nested: 用于嵌套结构的验证
// group: 用于分组验证

func CheckSQLInjection(obj interface{}) error {
	return checkStruct(reflect.ValueOf(obj), "")
}

func checkStruct(v reflect.Value, parent string) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 是否跳过字段检查
		if fieldType.Tag.Get("checksql") == "false" {
			continue
		}

		fieldName := fieldType.Name
		if parent != "" {
			fieldName = parent + "." + fieldName
		}

		switch field.Kind() {
		case reflect.String:
			if hasSQLInjection(field.String()) {
				return fmt.Errorf("字段 [%s] 存在非法字符", fieldName)
			}
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.String {
				for j := 0; j < field.Len(); j++ {
					if hasSQLInjection(field.Index(j).String()) {
						return fmt.Errorf("字段 [%s][%d] 存在非法字符", fieldName, j)
					}
				}
			}
		case reflect.Map:
			if field.Type().Key().Kind() == reflect.String && field.Type().Elem().Kind() == reflect.String {
				iter := field.MapRange()
				for iter.Next() {
					val := iter.Value().String()
					if hasSQLInjection(val) {
						return fmt.Errorf("字段 [%s][%s] 存在非法字符", fieldName, iter.Key().String())
					}
				}
			}
		case reflect.Struct:
			// 递归检查嵌套结构体
			if err := checkStruct(field, fieldName); err != nil {
				return err
			}
		case reflect.Ptr:
			if field.Elem().Kind() == reflect.Struct {
				if err := checkStruct(field, fieldName); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func hasSQLInjection(input string) bool {
	lower := strings.ToLower(input)
	keywords := []string{
		"'", "\"", ";", "--", "/*", "*/",
		"select ", "insert ", "update ", "delete ",
		"drop ", "exec ", "union ", " or ", " and ",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
