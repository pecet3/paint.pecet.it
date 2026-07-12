package gentstypes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFileAndGenerateTS(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gentstypes_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goFileContent := `
package models

// User represents a user profile
// name: CustomUser
type User struct {
	ID        int       ` + "`" + `json:"id"` + "`" + ` // Unique identifier
	Name      string    ` + "`" + `json:"name"` + "`" + `
	IsActive  bool      ` + "`" + `json:"is_active"` + "`" + `
	MetaData  interface{} ` + "`" + `json:"metadata"` + "`" + `
	Role      Role      ` + "`" + `json:"role"` + "`" + `
	Tags      []string  ` + "`" + `json:"tags"` + "`" + `
}

type Role struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

type IgnoredStruct struct {
	Secret string
}
`
	inputPath := filepath.Join(tmpDir, "models.go")
	err = os.WriteFile(inputPath, []byte(goFileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "output", "types.ts")

	Exec(tmpDir, outputPath)

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not generated: %s", outputPath)
	}

	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	outputStr := string(outputBytes)

	if !strings.Contains(outputStr, "export type CustomUser = {") {
		t.Errorf("Expected 'CustomUser' type definition missing or incorrect")
	}
	if !strings.Contains(outputStr, "id: number;") {
		t.Errorf("Expected field 'id: number;' missing")
	}
	if !strings.Contains(outputStr, "name: string;") {
		t.Errorf("Expected field 'name: string;' missing")
	}
	if !strings.Contains(outputStr, "is_active: boolean;") {
		t.Errorf("Expected field 'is_active: boolean;' missing")
	}
	if !strings.Contains(outputStr, "metadata: any;") {
		t.Errorf("Expected field 'metadata: any;' missing")
	}
	if !strings.Contains(outputStr, "tags: string[];") {
		t.Errorf("Expected slice mapping 'tags: string[];' missing")
	}

	if !strings.Contains(outputStr, "role: Role;") {
		t.Errorf("Expected nested struct reference 'role: Role;' missing")
	}

	if !strings.Contains(outputStr, "// User represents a user profile") {
		t.Errorf("Expected struct level comment missing")
	}
	if !strings.Contains(outputStr, "// Unique identifier") {
		t.Errorf("Expected field level comment missing")
	}

	if strings.Contains(outputStr, "IgnoredStruct") {
		t.Errorf("Struct without json tags should not be exported")
	}
}

func TestMapBasicType(t *testing.T) {
	tests := []struct {
		goType string
		want   string
	}{
		{"int", "number"},
		{"int64", "number"},
		{"float64", "number"},
		{"bool", "boolean"},
		{"string", "string"},
		{"time.Time", "string"},
		{"interface{}", "any"},
		{"unknownType", "any"},
	}

	for _, tt := range tests {
		t.Run(tt.goType, func(t *testing.T) {
			got := mapBasicType(tt.goType)
			if got != tt.want {
				t.Errorf("mapBasicType(%q) = %q; want %q", tt.goType, got, tt.want)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user", "User"},
		{"User", "User"},
		{"", ""},
		{"a", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := capitalize(tt.input)
			if got != tt.want {
				t.Errorf("capitalize(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractJSONName(t *testing.T) {
	tests := []struct {
		tag  string
		want string
	}{
		{"`json:\"id\"`", "id"},
		{"`json:\"user_id,omitempty\"`", "user_id"},
		{"`json:\"-\"`", "-"},
		{"`gorm:\"primaryKey\"`", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := extractJSONName(tt.tag)
			if got != tt.want {
				t.Errorf("extractJSONName(%q) = %q; want %q", tt.tag, got, tt.want)
			}
		})
	}
}
