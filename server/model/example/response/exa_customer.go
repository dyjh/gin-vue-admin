package response

import "github.com/dyjh/order-food-mini-app/server/model/example"

type ExaCustomerResponse struct {
	Customer example.ExaCustomer `json:"customer"`
}
