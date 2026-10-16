package timer

import "testing"

// TestGetErrorHandlerScript_Empty 未配置时应返回空串。
func TestGetErrorHandlerScript_Empty(t *testing.T) {
	v := getErrorHandlerScript()
	if v != "" {
		t.Fatalf("期望空串, got %q", v)
	}
}
