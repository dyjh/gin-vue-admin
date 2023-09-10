package utils

import (
	"fmt"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

func Validate(form interface{}) (validateErr string) {
	validateErr = ""
	validate := validator.New()
	errs := validate.Struct(form)
	trans := validateTransInit(validate)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErr = err.Translate(trans)
			break
		}
	}
	return validateErr
}

func validateTransInit(validate *validator.Validate) ut.Translator {
	// 万能翻译器，保存所有的语言环境和翻译数据
	uni := ut.New(zh.New())
	// 翻译器
	trans, _ := uni.GetTranslator("zh")
	//验证器注册翻译器
	err := zhTranslations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		fmt.Println(err)
	}
	return trans
}
