package initialize

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

func ValidatorInit() (*validator.Validate, ut.Translator, error) {
	translator := zh.New()
	uni := ut.New(translator, translator)

	trans, ok := uni.GetTranslator("zh")
	if !ok {
		return nil, nil, fmt.Errorf("translator zh not found")
	}

	validate, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return nil, nil, fmt.Errorf("gin binding validator engine is not *validator.Validate")
	}

	phoneValidator := func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		if len(phone) != 11 {
			return false
		}
		return phoneRegexp.MatchString(phone)
	}

	if err := validate.RegisterValidation("phone", phoneValidator); err != nil {
		return nil, nil, err
	}

	if err := zh_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		return nil, nil, err
	}

	if err := validate.RegisterTranslation("phone", trans, func(ut ut.Translator) error {
		return ut.Add("phone", "{0}必须是有效的手机号", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("phone", fe.Field())
		return t
	}); err != nil {
		return nil, nil, err
	}

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return validate, trans, nil
}
