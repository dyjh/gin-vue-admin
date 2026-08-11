package common

import (
	"encoding/json"
	"strings"

	appErrors "github.com/dyjh/order-food-mini-app/server/errors"
	"gorm.io/datatypes"
)

func redactAuditValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			normalized := strings.ToLower(key)
			if strings.Contains(normalized, "credential") ||
				normalized == "systemprompt" ||
				normalized == "userprompttemplate" ||
				normalized == "renderedsystemprompt" ||
				normalized == "rendereduserprompt" {
				result[key] = "[REDACTED]"
				continue
			}
			result[key] = redactAuditValue(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(typed))
		for index := range typed {
			result[index] = redactAuditValue(typed[index])
		}
		return result
	default:
		return value
	}
}

// SafeAuditJSON serializes an audit snapshot after redacting sensitive AI fields.
func SafeAuditJSON(value interface{}) (datatypes.JSON, error) {
	if value == nil {
		return datatypes.JSON([]byte("null")), nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "encode AI audit summary")
	}
	var generic interface{}
	if err := json.Unmarshal(encoded, &generic); err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "normalize AI audit summary")
	}
	safeEncoded, err := json.Marshal(redactAuditValue(generic))
	if err != nil {
		return nil, appErrors.AdminInternal.Wrap(err, "encode safe AI audit summary")
	}
	return datatypes.JSON(safeEncoded), nil
}
