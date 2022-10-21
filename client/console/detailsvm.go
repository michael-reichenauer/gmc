package console

import (
	"fmt"
	"strings"

	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/samber/lo"
)

type detailsVM struct {
	view cui.View
	text string
}

func NewDetailsVM(view cui.View) *detailsVM {
	return &detailsVM{view: view}
}

func (t *detailsVM) getCommitDetails(viewPort cui.ViewPage) (string, error) {
	return t.text, nil
}

func (t *detailsVM) setCurrentLine(line int, repo api.Repo, repoId string, ap api.Api) {
	log.Infof("line %d", line)
	t.text = "..."

	c := repo.Commits[line]
	cb := repo.Branches[c.BranchIndex]

	var cd api.CommitDetailsRsp
	err := ap.GetCommitDetails(api.CommitDetailsReq{RepoID: repoId, CommitID: c.ID}, &cd)
	if err != nil {
		log.Warnf("Failed: %v", err)
		return
	}

	_, message := getDetails(c, cb, cd)
	t.text = message
	t.view.PostOnUIThread(func() { t.view.NotifyChanged() })

}

func getDetails(c api.Commit, cb api.Branch, cd api.CommitDetailsRsp) (string, string) {

	files := strings.Join(cd.Files, "\n")
	id := c.ID

	parents := lo.Map(c.ParentIDs, func(v string, _ int) string { return v[:6] })
	children := lo.Map(c.ChildIDs, func(v string, _ int) string {
		if v == git.UncommittedID {
			return "uncommitted"
		} else {
			return v[:6]
		}
	})
	sid := c.SID

	if c.ID == git.UncommittedID {
		id = ""
		sid = "uncommitted"
	}
	title := "Commit: " + sid

	remote := ""
	if c.IsLocalOnly {
		remote = cui.Dark("Remote:   ") + cui.GreenDk("▲") + " pushable\n"
	}
	if c.IsRemoteOnly {
		remote = cui.Dark("Remote:   ") + cui.Blue("▼") + " pullable\n"
	}

	message := fmt.Sprintf(
		cui.Dark("Id:")+"       %s\n"+
			cui.Dark("Branch:  ")+" %s\n"+
			cui.Dark("Children:")+" %s\n"+
			cui.Dark("Parents: ")+" %s\n"+
			remote+
			"%s\n\n"+
			cui.Blue(strings.Repeat("_", 50))+
			cui.Blue("\n%d Files:\n")+
			"%s",
		id, cui.ColorText(cui.Color(cb.Color), cb.Name),
		strings.Join(children, ", "), strings.Join(parents, ", "),
		cd.Message, len(cd.Files), files)

	return title, message
}
