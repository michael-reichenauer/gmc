package git

import (
	"fmt"
	"gmc/utils/log"
	"os/exec"
	"strings"
	"time"
)

const (
	customRFC3339 = "2006-01-02T15:04:05Z0700" // Almost RFC3339 but no ':' in the last 4 chars
)

func getLog(path string) ([]Commit, error) {
	logText, err := getGitLog(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get git log, %v", err)
	}

	// Parse the git output lines into git commits
	return parseCommits(logText)
}

func getGitLog(path string) (string, error) {
	cmd := exec.Command("git", "log", "--all", "--pretty=%H|%ai|%ci|%an|%P|%s")
	cmd.Dir = path

	// Get the git log output
	out, err := cmd.Output()
	if err != nil {
		log.Warnf("Failed %v", err)
		return "", fmt.Errorf("failed to get git log, %v", err)
	}
	return string(out), nil
}

func parseCommits(logText string) ([]Commit, error) {
	var commits []Commit
	lines := strings.Split(logText, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		commit, err := parseCommit(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse git log output, %v", err)
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func parseCommit(line string) (Commit, error) {
	lineParts := strings.Split(line, "|")
	if len(lineParts) < 6 {
		return Commit{}, fmt.Errorf("failed to parse git commit %q", line)
	}
	subject := parseSubject(lineParts)
	id := lineParts[0]
	message := subject
	author := lineParts[3]
	parentIDs := parseParentIDs(lineParts)

	authorTime, commitTime, err := parseCommitTimes(lineParts)
	if err != nil {
		return Commit{}, fmt.Errorf("failed to parse commit times from commit %q, %v", line, err)
	}
	return Commit{ID: id, SID: id[:6], ParentIDs: parentIDs, Subject: subject,
		Message: message, Author: author, AuthorTime: authorTime, CommitTime: commitTime}, nil
}

func parseCommitTimes(lineParts []string) (time.Time, time.Time, error) {
	authorTime, err := time.Parse(customRFC3339, toCustomRFC3339Text(lineParts[1]))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse commit author time, %q, %v", lineParts[1], err)
	}
	commitTime, err := time.Parse(customRFC3339, toCustomRFC3339Text(lineParts[2]))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse commit commit time, %q, %v", lineParts[2], err)
	}
	return authorTime, commitTime, nil
}

func parseParentIDs(lineParts []string) []string {
	if lineParts[4] == "" {
		// No parents, (root commit has no parent)
		return nil
	}
	return strings.Split(lineParts[4], " ")
}

func toCustomRFC3339Text(gitTimeText string) string {
	timeText := strings.Replace(gitTimeText, " ", "T", 1)
	timeText = strings.Replace(timeText, " -", "-", 1)
	return strings.Replace(timeText, " +", "+", 1)
}
func parseSubject(lineParts []string) string {
	if len(lineParts) > 6 {
		// The subject contains one or more "|", so rejoin these parts into original subject
		return strings.Join(lineParts[5:], "|")
	}
	return lineParts[5]
}
