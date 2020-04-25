package git

import (
	"fmt"
	"strings"
	"time"
)

type Commit struct {
	ID         string
	SID        string
	ParentIDs  []string
	Subject    string
	Message    string
	Author     string
	AuthorTime time.Time
	CommitTime time.Time
}

func (c *Commit) String() string {
	return fmt.Sprintf("%s %s", c.SID, c.Subject)
}

const (
	customRFC3339 = "2006-01-02T15:04:05Z0700" // Almost RFC3339 but no ':' in the last 4 chars
)

type logService struct {
	cmd gitCommander
}

func ToSid(commitID string) string {
	return commitID[:6]
}

func newLog(cmd gitCommander) *logService {
	return &logService{cmd: cmd}
}

func (t *logService) getLog() ([]Commit, error) {
	logText, err := t.cmd.Git("log", "--all", "-z", "--pretty=%H|%ai|%ci|%an|%P|%B")
	if err != nil {
		return nil, fmt.Errorf("failed to get git log, %v", err)
	}

	// Parse the git output lines into git commits
	return t.parseCommits(logText)
}

func (t *logService) parseCommits(logText string) ([]Commit, error) {
	var commits []Commit
	lines := strings.Split(logText, "\x00")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		commit, err := t.parseCommit(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse git log output, %v", err)
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func (t *logService) parseCommit(line string) (Commit, error) {
	lineParts := strings.Split(line, "|")
	if len(lineParts) < 6 {
		return Commit{}, fmt.Errorf("failed to parse git commit %q", line)
	}
	subject, message := t.parseMessage(lineParts)
	id := lineParts[0]
	author := lineParts[3]
	parentIDs := t.parseParentIDs(lineParts)

	authorTime, commitTime, err := t.parseCommitTimes(lineParts)
	if err != nil {
		return Commit{}, fmt.Errorf("failed to parse commit times from commit %q, %v", line, err)
	}
	return Commit{ID: id, SID: ToSid(id), ParentIDs: parentIDs, Subject: subject,
		Message: message, Author: author, AuthorTime: authorTime, CommitTime: commitTime}, nil
}

func (t *logService) parseCommitTimes(lineParts []string) (time.Time, time.Time, error) {
	authorTime, err := time.Parse(customRFC3339, t.toCustomRFC3339Text(lineParts[1]))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse commit author time, %q, %v", lineParts[1], err)
	}
	commitTime, err := time.Parse(customRFC3339, t.toCustomRFC3339Text(lineParts[2]))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse commit commit time, %q, %v", lineParts[2], err)
	}
	return authorTime, commitTime, nil
}

func (t *logService) parseParentIDs(lineParts []string) []string {
	if lineParts[4] == "" {
		// No parents, (root commit has no parent)
		return nil
	}
	return strings.Split(lineParts[4], " ")
}

func (t *logService) toCustomRFC3339Text(gitTimeText string) string {
	timeText := strings.Replace(gitTimeText, " ", "T", 1)
	timeText = strings.Replace(timeText, " -", "-", 1)
	return strings.Replace(timeText, " +", "+", 1)
}

func (t *logService) parseMessage(lineParts []string) (string, string) {
	message := lineParts[5]
	if len(lineParts) > 6 {
		// The subject contains one or more "|", so rejoin these parts into original subject
		message = strings.Join(lineParts[5:], "|")
	}
	lines := strings.Split(message, "\n")
	return lines[0], message
}
