package gitlib

import (
	"fmt"
	"strings"
	"time"
)

const (
	customRFC3339 = "2006-01-02T15:04:05Z0700" // Almost RFC3339 but no ':' in the last 4 chars
)

type logHandler struct {
	cmd GitCommander
}

func newLog(cmd GitCommander) *logHandler {
	return &logHandler{cmd: cmd}
}

func (h *logHandler) getLog() ([]Commit, error) {
	logText, err := h.cmd.Git("log", "--all", "-z", "--pretty=%H|%ai|%ci|%an|%P|%B")
	if err != nil {
		return nil, fmt.Errorf("failed to get git log, %v", err)
	}

	// Parse the git output lines into git commits
	return h.parseCommits(logText)
}

func (h *logHandler) parseCommits(logText string) ([]Commit, error) {
	var commits []Commit
	lines := strings.Split(logText, "\x00")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		commit, err := h.parseCommit(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse git log output, %v", err)
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func (h *logHandler) parseCommit(line string) (Commit, error) {
	lineParts := strings.Split(line, "|")
	if len(lineParts) < 6 {
		return Commit{}, fmt.Errorf("failed to parse git commit %q", line)
	}
	subject, message := h.parseMessage(lineParts)
	id := lineParts[0]
	author := lineParts[3]
	parentIDs := h.parseParentIDs(lineParts)

	authorTime, commitTime, err := h.parseCommitTimes(lineParts)
	if err != nil {
		return Commit{}, fmt.Errorf("failed to parse commit times from commit %q, %v", line, err)
	}
	return Commit{ID: id, SID: id[:6], ParentIDs: parentIDs, Subject: subject,
		Message: message, Author: author, AuthorTime: authorTime, CommitTime: commitTime}, nil
}

func (h *logHandler) parseCommitTimes(lineParts []string) (time.Time, time.Time, error) {
	authorTime, err := time.Parse(customRFC3339, h.toCustomRFC3339Text(lineParts[1]))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse commit author time, %q, %v", lineParts[1], err)
	}
	commitTime, err := time.Parse(customRFC3339, h.toCustomRFC3339Text(lineParts[2]))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse commit commit time, %q, %v", lineParts[2], err)
	}
	return authorTime, commitTime, nil
}

func (h *logHandler) parseParentIDs(lineParts []string) []string {
	if lineParts[4] == "" {
		// No parents, (root commit has no parent)
		return nil
	}
	return strings.Split(lineParts[4], " ")
}

func (h *logHandler) toCustomRFC3339Text(gitTimeText string) string {
	timeText := strings.Replace(gitTimeText, " ", "T", 1)
	timeText = strings.Replace(timeText, " -", "-", 1)
	return strings.Replace(timeText, " +", "+", 1)
}

func (h *logHandler) parseMessage(lineParts []string) (string, string) {
	message := lineParts[5]
	if len(lineParts) > 6 {
		// The subject contains one or more "|", so rejoin these parts into original subject
		message = strings.Join(lineParts[5:], "|")
	}
	lines := strings.Split(message, "\n")
	return lines[0], message
}
