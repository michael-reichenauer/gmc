package viewrepo

import (
	"github.com/michael-reichenauer/gmc/api"
)

func toViewRepo(repo *viewRepo) api.VRepo {
	return api.VRepo{
		Commits:            toCommits(repo),
		CurrentBranchName:  repo.CurrentBranchName,
		RepoPath:           repo.WorkingFolder,
		UncommittedChanges: repo.UncommittedChanges,
		MergeMessage:       repo.MergeMessage,
		Conflicts:          repo.Conflicts,
	}
}

func toCommits(repo *viewRepo) []api.Commit {
	commits := make([]api.Commit, len(repo.Commits))
	for i, c := range repo.Commits {
		commits[i] = toCommit(c)
	}
	return commits
}

func toCommit(c *commit) api.Commit {
	return api.Commit{
		ID:           c.ID,
		SID:          c.SID,
		Subject:      c.Subject,
		Message:      c.Message,
		ParentIDs:    c.ParentIDs,
		ChildIDs:     c.ChildIDs,
		Author:       c.Author,
		AuthorTime:   c.AuthorTime,
		IsCurrent:    c.IsCurrent,
		Branch:       toBranch(c.Branch),
		Graph:        c.graph,
		More:         c.More,
		BranchTips:   c.BranchTips,
		IsLocalOnly:  c.IsLocalOnly,
		IsRemoteOnly: c.IsRemoteOnly,
		Tags:         c.Tags,
	}
}

func toBranch(b *branch) api.Branch {
	return api.Branch{
		Name:          b.name,
		DisplayName:   b.displayName,
		Index:         b.index,
		IsMultiBranch: b.isMultiBranch,
		RemoteName:    b.remoteName,
		LocalName:     b.localName,
		IsRemote:      b.isRemote,
		IsGitBranch:   b.isGitBranch,
		TipID:         b.tipId,
		IsCurrent:     b.isCurrent,
		HasRemoteOnly: b.HasRemoteOnly,
		HasLocalOnly:  b.HasLocalOnly,
		Color:         api.Color(b.color),
	}
}
