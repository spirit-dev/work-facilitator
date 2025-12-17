/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package helper

import (
	"regexp"

	"github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
)

// FileMatchesIgnorePattern checks if a file path matches any of the provided compiled regex patterns
func FileMatchesIgnorePattern(filePath string, patterns []*regexp.Regexp) bool {
	for _, re := range patterns {
		if re.MatchString(filePath) {
			log.Debugf("File '%s' matches ignore pattern '%s'\n", filePath, re.String())
			return true
		}
	}
	return false
}

// FilterIgnoredFiles returns only files that DON'T match any ignore patterns
func FilterIgnoredFiles(files []string, patterns []*regexp.Regexp) []string {
	var filtered []string
	for _, file := range files {
		if !FileMatchesIgnorePattern(file, patterns) {
			filtered = append(filtered, file)
		} else {
			log.Debugf("Filtering out ignored file: %s\n", file)
		}
	}
	return filtered
}

// GetStagedIgnoredFiles returns staged files that match any ignore pattern
func GetStagedIgnoredFiles(status git.Status, patterns []*regexp.Regexp) []string {
	var stagedIgnored []string

	for filePath, fileStatus := range status {
		// Check if file is staged (Added, Modified, or Renamed in staging area)
		if fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked {
			if FileMatchesIgnorePattern(filePath, patterns) {
				stagedIgnored = append(stagedIgnored, filePath)
			}
		}
	}

	return stagedIgnored
}

// UnstageIgnoredFiles is deprecated and should not be used.
// Instead, RepoCommit will fail if ignored files are staged.
func UnstageIgnoredFiles(patterns []*regexp.Regexp) error {
	// This function is no longer used but kept for backward compatibility
	return nil
}
