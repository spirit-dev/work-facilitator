/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package helper

import (
	"testing"
)

func TestTestStandard(t *testing.T) {
	tests := []struct {
		name            string
		commit          string
		commitExpr      string
		branch          string
		branchExpr      string
		standardEnforced bool
		expected        bool
	}{
		{
			name:             "enforced with matching commit and branch",
			commit:           "feat: ",
			commitExpr:       `^(feat|fix):.*`,
			branch:           "feat/test-branch",
			branchExpr:       `^feat/.*`,
			standardEnforced: true,
			expected:         true,
		},
		{
			name:             "enforced with non-matching commit",
			commit:           "bad: ",
			commitExpr:       `^(feat|fix):.*`,
			branch:           "feat/test-branch",
			branchExpr:       `^feat/.*`,
			standardEnforced: true,
			expected:         false,
		},
		{
			name:             "enforced with non-matching branch",
			commit:           "feat: ",
			commitExpr:       `^(feat|fix):.*`,
			branch:           "bad-branch",
			branchExpr:       `^feat/.*`,
			standardEnforced: true,
			expected:         false,
		},
		{
			name:             "not enforced accepts anything",
			commit:           "anything",
			commitExpr:       `^(feat|fix):.*`,
			branch:           "anything",
			branchExpr:       `^feat/.*`,
			standardEnforced: false,
			expected:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TestStandard(tt.commit, tt.commitExpr, tt.branch, tt.branchExpr, tt.standardEnforced)
			if result != tt.expected {
				t.Errorf("TestStandard() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCleanString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		separator string
		expected  string
	}{
		{
			name:      "alphanumeric only",
			input:     "Hello World",
			separator: "-",
			expected:  "Hello-World",
		},
		{
			name:      "with special characters",
			input:     "Hello! @World#",
			separator: "-",
			expected:  "Hello-World",
		},
		{
			name:      "multiple separators merged",
			input:     "Hello!!! World",
			separator: "-",
			expected:  "Hello-World",
		},
		{
			name:      "trailing separators removed",
			input:     "Hello World!",
			separator: "-",
			expected:  "Hello-World",
		},
		{
			name:      "empty input",
			input:     "",
			separator: "-",
			expected:  "",
		},
		{
			name:      "underscore separator",
			input:     "Feature Branch Name",
			separator: "_",
			expected:  "Feature_Branch_Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanString(tt.input, tt.separator)
			if result != tt.expected {
				t.Errorf("CleanString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCleanGlabString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Draft prefix removed",
			input:    "Draft: Resolve \"bug in login\"",
			expected: "bug in login",
		},
		{
			name:     "No draft prefix",
			input:    "Fix login issue",
			expected: "Fix login issue",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanGlabString(tt.input)
			if result != tt.expected {
				t.Errorf("CleanGlabString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "simple substitution",
			template: "Hello {{name}}",
			data:     map[string]interface{}{"name": "World"},
			expected: "Hello World",
		},
		{
			name:     "multiple substitutions",
			template: "{{greeting}} {{name}}!",
			data:     map[string]interface{}{"greeting": "Hello", "name": "World"},
			expected: "Hello World!",
		},
		{
			name:     "no placeholders",
			template: "Hello World",
			data:     map[string]interface{}{},
			expected: "Hello World",
		},
		{
			name:     "missing data replaced with empty",
			template: "Hello {{name}}",
			data:     map[string]interface{}{},
			expected: "Hello ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Template(tt.template, tt.data)
			if result != tt.expected {
				t.Errorf("Template() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefineCommit(t *testing.T) {
	tests := []struct {
		name        string
		branchType  string
		typeMapping string
		expected    string
	}{
		{
			name:        "feature branch maps to feat",
			branchType:  "feature",
			typeMapping: `{"feature":"feat","bugfix":"fix","hotfix":"fix"}`,
			expected:    "feat",
		},
		{
			name:        "bugfix branch maps to fix",
			branchType:  "bugfix",
			typeMapping: `{"feature":"feat","bugfix":"fix","hotfix":"fix"}`,
			expected:    "fix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefineCommit(tt.branchType, tt.typeMapping)
			if result != tt.expected {
				t.Errorf("DefineCommit() = %v, want %v", result, tt.expected)
			}
		})
	}
}
