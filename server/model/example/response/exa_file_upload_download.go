package response

import "github.com/dyjh/order-food-mini-app/server/model/example"

type ExaFileResponse struct {
	File example.ExaFileUploadAndDownload `json:"file"`
}
