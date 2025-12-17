/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package helper

import (
	"regexp"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestFileMatchesIgnorePattern_MatchesPattern(t *testing.T) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`out\.ya?ml$`),
		regexp.MustCompile(`out\d+\.ya?ml$`),
	}

	testCases := []struct {
		filePath string
		expected bool
	}{
		{"out.yaml", true},
		{"out.yml", true},
		{"out123.yaml", true},
		{"out456.yml", true},
		{"out1.yaml", true},
		{"config.yaml", false},
		{"output.yaml", false},
		{"out.txt", false},
		{"src/out.yaml", true},
		{"path/to/out123.yml", true},
	}

	for _, tc := range testCases {
		result := FileMatchesIgnorePattern(tc.filePath, patterns)
		if result != tc.expected {
			t.Errorf("FileMatchesIgnorePattern(%q) = %v; expected %v", tc.filePath, result, tc.expected)
		}
	}
}

func TestFileMatchesIgnorePattern_NoMatch(t *testing.T) {
	patterns := []*regexp.Regexp{regexp.MustCompile(`out\.ya?ml$`)}
	filePath := "config.yaml"

	result := FileMatchesIgnorePattern(filePath, patterns)
	if result {
		t.Errorf("FileMatchesIgnorePattern(%q) = true; expected false", filePath)
	}
}

func TestFileMatchesIgnorePattern_EmptyPatterns(t *testing.T) {
	patterns := []*regexp.Regexp{}
	filePath := "out.yaml"

	result := FileMatchesIgnorePattern(filePath, patterns)
	if result {
		t.Errorf("FileMatchesIgnorePattern(%q) with empty patterns = true; expected false", filePath)
	}
}

func TestFilterIgnoredFiles_FiltersCorrectly(t *testing.T) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`out\.ya?ml$`),
		regexp.MustCompile(`out\d+\.ya?ml$`),
	}
	files := []string{
		"src/main.go",
		"out.yaml",
		"README.md",
		"out123.yml",
		"config.yaml",
	}

	expected := []string{
		"src/main.go",
		"README.md",
		"config.yaml",
	}

	result := FilterIgnoredFiles(files, patterns)

	if len(result) != len(expected) {
		t.Errorf("FilterIgnoredFiles() returned %d files; expected %d", len(result), len(expected))
	}

	for i, file := range result {
		if file != expected[i] {
			t.Errorf("FilterIgnoredFiles()[%d] = %q; expected %q", i, file, expected[i])
		}
	}
}

func TestFilterIgnoredFiles_EmptyList(t *testing.T) {
	patterns := []*regexp.Regexp{regexp.MustCompile(`out\.ya?ml$`)}
	files := []string{}

	result := FilterIgnoredFiles(files, patterns)

	if len(result) != 0 {
		t.Errorf("FilterIgnoredFiles() with empty list returned %d files; expected 0", len(result))
	}
}

func TestFilterIgnoredFiles_AllFiltered(t *testing.T) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`out\.ya?ml$`),
		regexp.MustCompile(`out\d+\.ya?ml$`),
	}
	files := []string{
		"out.yaml",
		"out123.yml",
		"out1.yaml",
	}

	result := FilterIgnoredFiles(files, patterns)

	if len(result) != 0 {
		t.Errorf("FilterIgnoredFiles() returned %d files; expected 0 (all should be filtered)", len(result))
	}
}

func TestGetStagedIgnoredFiles_FindsStagedIgnored(t *testing.T) {
	patterns := []*regexp.Regexp{regexp.MustCompile(`out\.ya?ml$`)}

	// Create a mock git.Status
	status := git.Status{
		"out.yaml":    &git.FileStatus{Staging: git.Added, Worktree: git.Unmodified},
		"src/main.go": &git.FileStatus{Staging: git.Modified, Worktree: git.Unmodified},
		"README.md":   &git.FileStatus{Staging: git.Unmodified, Worktree: git.Modified},
	}

	result := GetStagedIgnoredFiles(status, patterns)

	if len(result) != 1 {
		t.Errorf("GetStagedIgnoredFiles() returned %d files; expected 1", len(result))
	}

	if len(result) > 0 && result[0] != "out.yaml" {
		t.Errorf("GetStagedIgnoredFiles()[0] = %q; expected 'out.yaml'", result[0])
	}
}

func TestGetStagedIgnoredFiles_NoStagedIgnored(t *testing.T) {
	patterns := []*regexp.Regexp{regexp.MustCompile(`out\.ya?ml$`)}

	// Create a mock git.Status with no staged ignored files
	status := git.Status{
		"src/main.go": &git.FileStatus{Staging: git.Modified, Worktree: git.Unmodified},
		"README.md":   &git.FileStatus{Staging: git.Unmodified, Worktree: git.Modified},
	}

	result := GetStagedIgnoredFiles(status, patterns)

	if len(result) != 0 {
		t.Errorf("GetStagedIgnoredFiles() returned %d files; expected 0", len(result))
	}
}

func TestGetStagedIgnoredFiles_IgnoresUnstagedFiles(t *testing.T) {
	patterns := []*regexp.Regexp{regexp.MustCompile(`out\.ya?ml$`)}

	// Create a mock git.Status where out.yaml is modified but NOT staged
	status := git.Status{
		"out.yaml":    &git.FileStatus{Staging: git.Unmodified, Worktree: git.Modified},
		"src/main.go": &git.FileStatus{Staging: git.Modified, Worktree: git.Unmodified},
	}

	result := GetStagedIgnoredFiles(status, patterns)

	if len(result) != 0 {
		t.Errorf("GetStagedIgnoredFiles() returned %d files; expected 0 (out.yaml is not staged)", len(result))
	}
}
