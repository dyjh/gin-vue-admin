package utils

import (
	"testing"
)

func TestGetJSONKeys(t *testing.T) {
	var jsonStr = `
	{
		"Name": "test",
		"TableName": "test",
		"TemplateID": "test",
		"TemplateInfo": "test",
		"Limit": 0
}`
	keys, err := GetJSONKeys(jsonStr)
	if err != nil {
		t.Fatalf("GetJSONKeys failed: %v", err)
	}
	if len(keys) != 5 {
		t.Fatalf("GetJSONKeys returned %d keys, want 5", len(keys))
	}
	if keys[0] != "Name" {
		t.Fatalf("keys[0] = %q, want %q", keys[0], "Name")
	}
	if keys[1] != "TableName" {
		t.Fatalf("keys[1] = %q, want %q", keys[1], "TableName")
	}
	if keys[2] != "TemplateID" {
		t.Fatalf("keys[2] = %q, want %q", keys[2], "TemplateID")
	}
	if keys[3] != "TemplateInfo" {
		t.Fatalf("keys[3] = %q, want %q", keys[3], "TemplateInfo")
	}
	if keys[4] != "Limit" {
		t.Fatalf("keys[4] = %q, want %q", keys[4], "Limit")
	}
}
