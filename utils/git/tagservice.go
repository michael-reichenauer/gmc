package git

import (
	"strings"
)

type Tag struct {
	CommitID string
	TagName  string
}

// fetch/push from remote origin
type tagService struct {
	cmd gitCommander
}

func newTagService(cmd gitCommander) *tagService {
	return &tagService{cmd: cmd}
}

func (t *tagService) getTags() ([]Tag, error) {
	output, err := t.cmd.Git("show-ref", "--dereference", "--tags")
	if err != nil {
		return []Tag{}, nil
	}
	tags := t.parseTags(output)
	return tags, nil
}

func (t *tagService) parseTags(output string) []Tag {
	var tags []Tag
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 51 {
			continue
		}
		commitID := line[0:40]
		name := line[51:]

		if strings.HasSuffix(name, "^{}") {
			// Seems that some client add a suffix for some reason
			name = name[:len(name)-3]
		}
		tags = append(tags, Tag{CommitID: commitID, TagName: name})
	}
	return tags
}
