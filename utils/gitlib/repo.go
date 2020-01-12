package gitlib

type Repo struct {
	cmd           GitCommander
	status        *statusHandler
	logHandler    *logHandler
	branches      *branchesHandler
	fetchService  *fetchService
	ignoreHandler *ignoreHandler
}

func NewRepo(path string) *Repo {
	cmd := newGitCmd(path)
	return &Repo{
		cmd:           cmd,
		status:        newStatus(cmd),
		logHandler:    newLog(cmd),
		branches:      newBranches(cmd),
		fetchService:  newFetch(cmd),
		ignoreHandler: newIgnoreHandler(path),
	}
}
func (h *Repo) RepoPath() string {
	return h.cmd.RepoPath()
}
func (h *Repo) GetLog() ([]Commit, error) {
	return h.logHandler.getLog()
}

func (h *Repo) GetBranches() ([]Branch, error) {
	return h.branches.getBranches()
}

func (h *Repo) GetStatus() (Status, error) {
	return h.status.getStatus()
}

func (h *Repo) Fetch() error {
	return h.fetchService.fetch()
}

func (h *Repo) IsIgnored(path string) bool {
	return h.ignoreHandler.isIgnored(path)
}
