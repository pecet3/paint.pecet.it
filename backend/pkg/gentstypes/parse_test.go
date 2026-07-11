package gentstypes

import (
	"reflect"
	"testing"
)

func TestParseRawField(t *testing.T) {

	tests := []struct {
		name     string
		raw      string
		expected goField
	}{
		{
			name: "simple",
			raw:  "Id int",
			expected: goField{
				goName: "Id",
				goType: "int",
			},
		},
		{
			name: "with tag",
			raw:  "Uuid string `json:\"uuid\"`",
			expected: goField{
				goName: "Uuid",
				goType: "string",
				tag:    "`json:\"uuid\"`",
			},
		},
		{
			name: "with inline comment",
			raw:  "Active bool // inline",
			expected: goField{
				goName:  "Active",
				goType:  "bool",
				comment: "inline",
			},
		},
		{
			name: "with prefixed comment",
			raw:  "COMMENT:prefixed|||Age int",
			expected: goField{
				goName:  "Age",
				goType:  "int",
				comment: "prefixed",
			},
		},
		{
			name: "with prefixed and inline comment",
			raw:  "COMMENT:prefixed|||Age int // inline",
			expected: goField{
				goName:  "Age",
				goType:  "int",
				comment: "prefixed\ninline",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRawField(tt.raw)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("got %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestBuildTSTypes(t *testing.T) {
	structsMap := map[string]*goStruct{
		"models.User": {
			packageName: "models",
			structName:  "User",
			fields: []goField{
				{goName: "Id", goType: "int", tag: "`json:\"id\"`"},
				{goName: "Name", goType: "string", tag: "`json:\"name\"`"},
				{goName: "IsActive", goType: "bool", tag: "`json:\"is_active\"`"},
				{goName: "Role", goType: "Role", tag: "`json:\"role\"`"},
				{goName: "Posts", goType: "[]Post", tag: "`json:\"posts\"`"},
				{goName: "Session", goType: "*auth.Session", tag: "`json:\"session\"`"},
				{goName: "Tags", goType: "[]string", tag: "`json:\"tags\"`"},
				{goName: "Unknown", goType: "magic.Type", tag: "`json:\"unknown\"`"},
				{goName: "Ignored", goType: "string", tag: "`json:\"-\"`"},
				{goName: "NoTag", goType: "string", tag: ""},
			},
		},
		"models.Role": {
			packageName: "models",
			structName:  "Role",
		},
		"models.Post": {
			packageName: "models",
			structName:  "Post",
		},
		"auth.Session": {
			packageName: "auth",
			structName:  "Session",
		},
	}

	tsTypes := buildTSTypes(structsMap)

	var userType tsType
	found := false
	for _, tsT := range tsTypes {
		if tsT.name == "ModelsUser" {
			userType = tsT
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("ModelsUser not found in generated types")
	}

	expectedFields := []tsField{
		{tsName: "id", tsType: "number"},
		{tsName: "name", tsType: "string"},
		{tsName: "is_active", tsType: "boolean"},
		{tsName: "role", tsType: "ModelsRole"},
		{tsName: "posts", tsType: "ModelsPost[]"},
		{tsName: "session", tsType: "AuthSession"},
		{tsName: "tags", tsType: "string[]"},
		{tsName: "unknown", tsType: "any"},
	}

	if len(userType.fields) != len(expectedFields) {
		t.Fatalf("expected %d fields, got %d", len(expectedFields), len(userType.fields))
	}

	for i, expected := range expectedFields {
		actual := userType.fields[i]
		if actual.tsName != expected.tsName || actual.tsType != expected.tsType {
			t.Errorf("field %d: expected %s: %s, got %s: %s", i, expected.tsName, expected.tsType, actual.tsName, actual.tsType)
		}
	}
}

func TestExtractJSONName(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{"`json:\"id\"`", "id"},
		{"`json:\"first_name,omitempty\"`", "first_name"},
		{"`json:\"-\"`", "-"},
		{"`xml:\"id\"`", ""},
		{"", ""},
	}

	for _, tt := range tests {
		result := extractJSONName(tt.tag)
		if result != tt.expected {
			t.Errorf("extractJSONName(%q) = %q, want %q", tt.tag, result, tt.expected)
		}
	}
}

func TestFormatTSName(t *testing.T) {
	tests := []struct {
		pkg      string
		str      string
		expected string
	}{
		{"models", "User", "ModelsUser"},
		{"auth", "Session", "AuthSession"},
		{"db", "connection", "DbConnection"},
	}

	for _, tt := range tests {
		result := formatTSName(tt.pkg, tt.str)
		if result != tt.expected {
			t.Errorf("formatTSName(%q, %q) = %q, want %q", tt.pkg, tt.str, result, tt.expected)
		}
	}
}
