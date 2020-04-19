package git

// fetches from remote origin
type mergeService struct {
	cmd GitCommander
}

func newMerge(cmd GitCommander) *mergeService {
	return &mergeService{cmd: cmd}
}

func (h *mergeService) mergeBranch(name string) error {
	// $"merge --no-ff --no-commit --stat --progress {name}", ct);
	_, err := h.cmd.Git("merge", "--no-ff", "-no-commit", "--stat", name)
	if err != nil {
		return err
	}
	return nil
}
