package augmented

import "github.com/michael-reichenauer/gmc/utils/git"

type Tag struct {
	CommitID string
	TagName  string
}

func toTags(gitTags []git.Tag) []Tag {
	tags := make([]Tag, len(gitTags))
	for i, tag := range gitTags {
		tags[i] = Tag{CommitID: tag.CommitID, TagName: tag.TagName}
	}
	return tags
}
