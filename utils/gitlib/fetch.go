package gitlib

type fetchService struct {
	cmd GitCommander
}

func newFetch(cmd GitCommander) *fetchService {
	return &fetchService{cmd: cmd}
}

func (h *fetchService) fetch() error {
	_, err := h.cmd.Git("fetch", "--prune", "--tags", "origin")
	if err != nil {
		return err
	}
	return nil
}
