package tools

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

func TestRegisterDocumentsAllowedTypeValues(t *testing.T) {
	validTaskTypes := []string{"feature", "bug"}
	validRelationTypes := []string{"blocked_by", "relates_to", "duplicate_of"}

	s := server.NewMCPServer("test-server", "1.0.0")
	Register(s, nil, validTaskTypes, validRelationTypes)

	tools := s.ListTools()

	assertStringProperty(t, tools["create_task"].Tool.InputSchema.Properties, "type",
		"Allowed values: feature, bug.",
		validTaskTypes,
	)
	assertStringProperty(t, tools["update_task"].Tool.InputSchema.Properties, "type",
		"Allowed values: feature, bug.",
		validTaskTypes,
	)
	assertStringProperty(t, tools["list_tasks"].Tool.InputSchema.Properties, "type",
		"Allowed values: feature, bug.",
		validTaskTypes,
	)
	assertStringProperty(t, tools["add_relation"].Tool.InputSchema.Properties, "type",
		"Allowed values: blocked_by, relates_to, duplicate_of.",
		validRelationTypes,
	)
	assertStringProperty(t, tools["remove_relation"].Tool.InputSchema.Properties, "type",
		"Allowed values: blocked_by, relates_to, duplicate_of.",
		validRelationTypes,
	)
}

func assertStringProperty(t *testing.T, properties map[string]any, name, wantDescriptionSuffix string, wantEnum []string) {
	t.Helper()

	raw, ok := properties[name]
	if !ok {
		t.Fatalf("property %q not found", name)
	}

	property, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("property %q has unexpected type %T", name, raw)
	}

	description, ok := property["description"].(string)
	if !ok {
		t.Fatalf("property %q description has unexpected type %T", name, property["description"])
	}
	if !strings.Contains(description, wantDescriptionSuffix) {
		t.Fatalf("property %q description = %q, want substring %q", name, description, wantDescriptionSuffix)
	}

	enumValues, ok := property["enum"].([]string)
	if !ok {
		t.Fatalf("property %q enum has unexpected type %T", name, property["enum"])
	}
	if len(enumValues) != len(wantEnum) {
		t.Fatalf("property %q enum length = %d, want %d", name, len(enumValues), len(wantEnum))
	}
	for i, value := range wantEnum {
		if enumValues[i] != value {
			t.Fatalf("property %q enum[%d] = %q, want %q", name, i, enumValues[i], value)
		}
	}
}
