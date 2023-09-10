package auth

type LoginForm struct {
	Code string `json:"code" validate:"required"`
	//UserName string `json:"user_name" validate:"required"`
	//Password string `json:"password" validate:"required,min=6,max=20"`
}
