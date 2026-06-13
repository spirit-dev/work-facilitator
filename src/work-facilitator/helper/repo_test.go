/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package helper

import (
	"strings"
	"testing"
)

func TestFormatUnifiedDiff_NoChanges(t *testing.T) {
	old := []string{"line1", "line2"}
	new := []string{"line1", "line2"}
	result := formatUnifiedDiff(old, new, 3)
	if result != "" {
		t.Errorf("Expected empty diff for identical content, got: %s", result)
	}
}

func TestFormatUnifiedDiff_SingleLineChange(t *testing.T) {
	old := []string{"line1", "line2", "line3", "line4", "line5"}
	new := []string{"line1", "line2_changed", "line3", "line4", "line5"}
	result := formatUnifiedDiff(old, new, 3)

	if !strings.Contains(result, "@@") {
		t.Errorf("Expected hunk header (@@), got: %s", result)
	}
	if !strings.Contains(result, "-line2") {
		t.Errorf("Expected removed line (-line2), got: %s", result)
	}
	if !strings.Contains(result, "+line2_changed") {
		t.Errorf("Expected added line (+line2_changed), got: %s", result)
	}
	// Context lines should be present
	if !strings.Contains(result, " line1") && !strings.Contains(result, " line3") {
		t.Errorf("Expected context lines (prefixed with space), got: %s", result)
	}
}

func TestFormatUnifiedDiff_Addition(t *testing.T) {
	old := []string{"line1", "line2"}
	new := []string{"line1", "line2", "line3"}
	result := formatUnifiedDiff(old, new, 3)

	if !strings.Contains(result, "@@") {
		t.Errorf("Expected hunk header, got: %s", result)
	}
	if !strings.Contains(result, "+line3") {
		t.Errorf("Expected added line, got: %s", result)
	}
}

func TestFormatUnifiedDiff_Deletion(t *testing.T) {
	old := []string{"line1", "line2", "line3"}
	new := []string{"line1", "line3"}
	result := formatUnifiedDiff(old, new, 3)

	if !strings.Contains(result, "@@") {
		t.Errorf("Expected hunk header, got: %s", result)
	}
	if !strings.Contains(result, "-line2") {
		t.Errorf("Expected removed line, got: %s", result)
	}
}

func TestFormatUnifiedDiff_MultipleHunks(t *testing.T) {
	old := []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
	}
	new := []string{
		"a", "b_CHANGED", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p_CHANGED", "q", "r", "s", "t",
	}
	result := formatUnifiedDiff(old, new, 3)

	// Should produce two hunks (changes far apart)
	hunkCount := strings.Count(result, "@@")
	if hunkCount < 2 {
		t.Errorf("Expected at least 2 hunks for changes far apart, got %d hunks. Result:\n%s", hunkCount, result)
	}
}

func TestFormatUnifiedDiff_ContextLines(t *testing.T) {
	old := []string{"a", "b", "c", "d", "e", "f", "g"}
	new := []string{"a", "b", "c", "d_CHANGED", "e", "f", "g"}
	result := formatUnifiedDiff(old, new, 2)

	lines := strings.Split(result, "\n")
	// Count context lines (lines starting with space, excluding hunk header)
	contextCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, " ") {
			contextCount++
		}
	}
	// With context=2, we expect up to 4 context lines (2 before + 2 after)
	if contextCount == 0 || contextCount > 4 {
		t.Errorf("Expected 1-4 context lines with context=2, got %d. Result:\n%s", contextCount, result)
	}
}

func TestFormatUnifiedDiff_EmptyOld(t *testing.T) {
	old := []string{}
	new := []string{"line1", "line2"}
	result := formatUnifiedDiff(old, new, 3)

	if !strings.Contains(result, "@@") {
		t.Errorf("Expected hunk header, got: %s", result)
	}
	if !strings.Contains(result, "+line1") {
		t.Errorf("Expected added lines, got: %s", result)
	}
}

func TestFormatUnifiedDiff_EmptyNew(t *testing.T) {
	old := []string{"line1", "line2"}
	new := []string{}
	result := formatUnifiedDiff(old, new, 3)

	if !strings.Contains(result, "@@") {
		t.Errorf("Expected hunk header, got: %s", result)
	}
	if !strings.Contains(result, "-line1") {
		t.Errorf("Expected removed lines, got: %s", result)
	}
}

func TestFormatUnifiedDiff_HunkHeaders(t *testing.T) {
	old := []string{"a", "b", "c", "d"}
	new := []string{"a", "b_CHANGED", "c", "d"}
	result := formatUnifiedDiff(old, new, 3)

	// Verify hunk header format: @@ -oldStart,oldCount +newStart,newCount @@
	if !strings.Contains(result, "@@ -") {
		t.Errorf("Expected hunk header with @@ - format, got: %s", result)
	}
	if !strings.Contains(result, " +") {
		t.Errorf("Expected hunk header with + count, got: %s", result)
	}
	if !strings.Contains(result, " @@") {
		t.Errorf("Expected hunk header closing, got: %s", result)
	}
}

func TestBuildHunks_BasicMerge(t *testing.T) {
	ops := []diffOp{
		{' ', "line1"},
		{'-', "line2"},
		{'+', "line2_changed"},
		{' ', "line3"},
	}
	hunks := buildHunks(ops, 3)
	if len(hunks) != 1 {
		t.Errorf("Expected 1 hunk for simple change, got %d", len(hunks))
	}
}

func TestBuildHunks_EmptyOps(t *testing.T) {
	hunks := buildHunks([]diffOp{}, 3)
	if len(hunks) != 0 {
		t.Errorf("Expected 0 hunks for empty ops, got %d", len(hunks))
	}
}

func TestComputeDiffOps_Identical(t *testing.T) {
	old := []string{"a", "b", "c"}
	new := []string{"a", "b", "c"}
	ops := computeDiffOps(old, new)

	for _, op := range ops {
		if op.action != ' ' {
			t.Errorf("Expected all keep ops for identical content, got: %c", op.action)
		}
	}
}

func TestComputeDiffOps_AllNew(t *testing.T) {
	old := []string{}
	new := []string{"a", "b", "c"}
	ops := computeDiffOps(old, new)

	for _, op := range ops {
		if op.action != '+' {
			t.Errorf("Expected all add ops for new content, got: %c", op.action)
		}
	}
	if len(ops) != 3 {
		t.Errorf("Expected 3 ops, got %d", len(ops))
	}
}

func TestComputeDiffOps_AllDeleted(t *testing.T) {
	old := []string{"a", "b", "c"}
	new := []string{}
	ops := computeDiffOps(old, new)

	for _, op := range ops {
		if op.action != '-' {
			t.Errorf("Expected all delete ops for deleted content, got: %c", op.action)
		}
	}
	if len(ops) != 3 {
		t.Errorf("Expected 3 ops, got %d", len(ops))
	}
}

func TestComputeDiffOps_Mixed(t *testing.T) {
	old := []string{"keep", "remove", "keep2"}
	new := []string{"keep", "add", "keep2"}
	ops := computeDiffOps(old, new)

	expected := []byte{' ', '-', '+', ' '}
	if len(ops) != len(expected) {
		t.Errorf("Expected %d ops, got %d", len(expected), len(ops))
		return
	}
	for i, op := range ops {
		if op.action != expected[i] {
			t.Errorf("Op %d: expected %c, got %c", i, expected[i], op.action)
		}
	}
}

func TestMergeRegions_Overlapping(t *testing.T) {
	regions := []region{
		{0, 5},
		{3, 10},
	}
	merged := mergeRegions(regions, 3)
	if len(merged) != 1 {
		t.Errorf("Expected 1 merged region, got %d", len(merged))
	}
	if merged[0].start != 0 || merged[0].end != 10 {
		t.Errorf("Expected merged region [0,10], got [%d,%d]", merged[0].start, merged[0].end)
	}
}

func TestMergeRegions_TooFar(t *testing.T) {
	regions := []region{
		{0, 5},
		{100, 105},
	}
	merged := mergeRegions(regions, 3)
	if len(merged) != 2 {
		t.Errorf("Expected 2 separate regions, got %d", len(merged))
	}
}
