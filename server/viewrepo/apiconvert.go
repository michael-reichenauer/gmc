package viewrepo

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/samber/lo"
)

func ToApiCommitDiffs(diff []git.CommitDiff) []api.CommitDiff {
	return lo.Map(diff, func(v git.CommitDiff, _ int) api.CommitDiff { return ToApiCommitDiff(v) })
}

func ToApiCommitDiff(diff git.CommitDiff) api.CommitDiff {
	return api.CommitDiff{
		Id:        diff.Id,
		Author:    diff.Author,
		Date:      diff.Date,
		Message:   diff.Message,
		FileDiffs: toApiFileDiffs(diff.FileDiffs),
	}
}

func toApiFileDiffs(gfd []git.FileDiff) []api.FileDiff {
	diffs := make([]api.FileDiff, len(gfd))
	for i, d := range gfd {
		diffs[i] = api.FileDiff{
			PathBefore:   d.PathBefore,
			PathAfter:    d.PathAfter,
			IsRenamed:    d.IsRenamed,
			DiffMode:     api.DiffMode(d.DiffMode),
			SectionDiffs: toApiSectionDiffs(d.SectionDiffs),
		}
	}
	return diffs
}

func toApiSectionDiffs(gsd []git.SectionDiff) []api.SectionDiff {
	diffs := make([]api.SectionDiff, len(gsd))
	for i, d := range gsd {
		diffs[i] = api.SectionDiff{
			ChangedIndexes: d.ChangedIndexes,
			LinesDiffs:     toApiLineDiffs(d.LinesDiffs),
		}
	}
	return diffs
}

func toApiLineDiffs(gld []git.LinesDiff) []api.LinesDiff {
	diffs := make([]api.LinesDiff, len(gld))
	for i, d := range gld {
		diffs[i] = api.LinesDiff{
			DiffMode: api.DiffMode(d.DiffMode),
			Line:     d.Line,
		}
	}
	return diffs
}

func toApiRepo(repo *repo) api.Repo {
	graph := toApiConsoleGraph(repo.Commits)
	return api.Repo{
		Commits:            toApiCommits(repo.Commits),
		Branches:           toApiBranches(repo.Branches),
		CurrentBranchName:  repo.CurrentBranchName,
		RepoPath:           repo.WorkingFolder,
		UncommittedChanges: repo.UncommittedChanges,
		MergeMessage:       repo.MergeMessage,
		Conflicts:          repo.Conflicts,
		ConsoleGraph:       graph,
	}
}

func toApiBranches(branches []*branch) []api.Branch {
	apiBranches := make([]api.Branch, len(branches))
	for i, b := range branches {
		apiBranches[i] = toApiBranch(b)
	}
	return apiBranches
}

func toApiCommits(commits []*commit) []api.Commit {
	apiCommits := make([]api.Commit, len(commits))
	for i, c := range commits {
		apiCommits[i] = toApiCommit(c)
	}
	return apiCommits
}

func toApiConsoleGraph(commits []*commit) api.Graph {
	rows := make([]api.GraphRow, len(commits))
	for i, c := range commits {
		rows[i] = c.graph
	}
	return rows
}

func toApiCommit(c *commit) api.Commit {
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

func toApiBranch(b *branch) api.Branch {
	return api.Branch{
		Name:                 b.name,
		DisplayName:          b.displayName,
		Index:                b.index,
		IsAmbiguousBranch:    b.isAmbiguousBranch,
		RemoteName:           b.remoteName,
		LocalName:            b.localName,
		IsRemote:             b.isRemote,
		IsGitBranch:          b.isGitBranch,
		IsMainBranch:         isMainBranch(b.name),
		TipID:                b.tipId,
		IsCurrent:            b.isCurrent,
		IsSetAsParent:        b.isSetAsParent,
		HasRemoteOnly:        b.HasRemoteOnly,
		HasLocalOnly:         b.HasLocalOnly,
		Color:                api.Color(b.color),
		AmbiguousBranchNames: b.ambiguousBranchNames,
	}
}
