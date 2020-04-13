package repoview

import (
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type commitVM struct {
}

func NewCommitVM() *commitVM {
	return &commitVM{}
}

func (h commitVM) getCommitDetails(viewPort ui.ViewPage) (string, error) {
	return "commit ", nil
}
