package viewrepo

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/git"
)

func ToCommitDiff(diff git.CommitDiff) api.CommitDiff {
	return api.CommitDiff{
		FileDiffs: toFileDiffs(diff.FileDiffs),
	}
}

func toFileDiffs(gfd []git.FileDiff) []api.FileDiff {
	diffs := make([]api.FileDiff, len(gfd))
	for i, d := range gfd {
		diffs[i] = api.FileDiff{
			PathBefore:   d.PathBefore,
			PathAfter:    d.PathAfter,
			IsRenamed:    d.IsRenamed,
			DiffMode:     api.DiffMode(d.DiffMode),
			SectionDiffs: toSectionDiffs(d.SectionDiffs),
		}
	}
	return diffs
}

func toSectionDiffs(gsd []git.SectionDiff) []api.SectionDiff {
	diffs := make([]api.SectionDiff, len(gsd))
	for i, d := range gsd {
		diffs[i] = api.SectionDiff{
			ChangedIndexes: d.ChangedIndexes,
			LinesDiffs:     toLineDiffs(d.LinesDiffs),
		}
	}
	return diffs
}

func toLineDiffs(gld []git.LinesDiff) []api.LinesDiff {
	diffs := make([]api.LinesDiff, len(gld))
	for i, d := range gld {
		diffs[i] = api.LinesDiff{
			DiffMode: api.DiffMode(d.DiffMode),
			Line:     d.Line,
		}
	}
	return diffs
}

func toViewRepo(repo *repo) api.Repo {
	graph := toConsoleGraph(repo)
	return api.Repo{
		Commits:            toCommits(repo),
		Branches:           toBranches(repo),
		CurrentBranchName:  repo.CurrentBranchName,
		RepoPath:           repo.WorkingFolder,
		UncommittedChanges: repo.UncommittedChanges,
		MergeMessage:       repo.MergeMessage,
		Conflicts:          repo.Conflicts,
		ConsoleGraph:       graph,
	}
}

func toBranches(repo *repo) []api.Branch {
	branches := make([]api.Branch, len(repo.Branches))
	for i, b := range repo.Branches {
		branches[i] = toBranch(b)
	}
	return branches
}

func toCommits(repo *repo) []api.Commit {
	commits := make([]api.Commit, len(repo.Commits))
	for i, c := range repo.Commits {
		commits[i] = toCommit(c)
	}
	return commits
}

func toConsoleGraph(repo *repo) api.Graph {
	rows := make([]api.GraphRow, len(repo.Commits))
	for i, c := range repo.Commits {
		rows[i] = c.graph
	}
	return rows
}

func toCommit(c *commit) api.Commit {
	return api.Commit{
		ID:                 c.ID,
		SID:                c.SID,
		Subject:            c.Subject,
		Message:            c.Message,
		ParentIDs:          c.ParentIDs,
		ChildIDs:           c.ChildIDs,
		Author:             c.Author,
		AuthorTime:         c.AuthorTime,
		IsCurrent:          c.IsCurrent,
		BranchIndex:        c.Branch.index,
		More:               c.More,
		BranchTips:         c.BranchTips,
		IsLocalOnly:        c.IsLocalOnly,
		IsRemoteOnly:       c.IsRemoteOnly,
		IsUncommitted:      c.ID == git.UncommittedID,
		IsPartialLogCommit: c.ID == git.PartialLogCommitID,
		Tags:               c.Tags,
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
