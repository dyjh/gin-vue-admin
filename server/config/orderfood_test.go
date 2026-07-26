package config

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestOrderFoodIdentityKeyIsNotSerialized 验证安全配置不会通过系统配置接口序列化。
func TestOrderFoodIdentityKeyIsNotSerialized(t *testing.T) {
	payload, err := json.Marshal(Server{
		OrderFood: OrderFood{IdentityKey: "sensitive-identity-key"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "sensitive-identity-key") ||
		strings.Contains(string(payload), "orderfood") {
		t.Fatalf("security config leaked into JSON: %s", payload)
	}
}
