package front

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type miniAppContract struct {
	BasePath   string `json:"basePath"`
	Interfaces []struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	} `json:"interfaces"`
}

func TestRegisterMatchesAllMiniAppContractRoutes(t *testing.T) {
	contractBytes, err := os.ReadFile("../../aiDoc/miniApp/api.json")
	if err != nil {
		t.Fatalf("read miniapp contract: %v", err)
	}
	var contract miniAppContract
	if err := json.Unmarshal(contractBytes, &contract); err != nil {
		t.Fatalf("decode miniapp contract: %v", err)
	}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	Register(engine.Group(""))

	actual := make(map[string]bool, len(engine.Routes()))
	for _, route := range engine.Routes() {
		actual[route.Method+" "+route.Path] = true
	}
	if len(actual) != len(contract.Interfaces) {
		t.Fatalf("registered %d routes, contract has %d", len(actual), len(contract.Interfaces))
	}
	for _, endpoint := range contract.Interfaces {
		path := contract.BasePath + endpoint.Path
		for {
			start := strings.Index(path, "{")
			if start < 0 {
				break
			}
			end := strings.Index(path[start:], "}")
			if end < 0 {
				t.Fatalf("invalid contract path %q", path)
			}
			end += start
			path = path[:start] + ":" + path[start+1:end] + path[end+1:]
		}
		key := endpoint.Method + " " + path
		if !actual[key] {
			t.Errorf("missing route %s", key)
		}
	}
}
