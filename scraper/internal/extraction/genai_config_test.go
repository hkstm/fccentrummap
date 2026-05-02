package extraction

import "testing"

func TestGeneratePass2RefinementContentConfigUsesRefinementFunction(t *testing.T) {
	cfg := GeneratePass2RefinementContentConfig()
	if cfg == nil || cfg.ToolConfig == nil || cfg.ToolConfig.FunctionCallingConfig == nil {
		t.Fatal("expected non-nil tool/function calling config")
	}
	allowed := cfg.ToolConfig.FunctionCallingConfig.AllowedFunctionNames
	if len(allowed) != 1 || allowed[0] != SubmitRefinedSpotsFunctionName {
		t.Fatalf("unexpected allowed functions: %+v", allowed)
	}
}
