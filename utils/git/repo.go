package git

type Repo struct {
	path string
}

func NewRepo(path string) *Repo {
	return &Repo{path: path}
}

func (r *Repo) GetLog() ([]Commit, error) {
	return getLog(r.path)
}

func (r *Repo) GetBranches() ([]Branch, error) {
	return getBranches(r.path)
}

func (r *Repo) GetStatus() (Status, error) {
	return getStatus(r.path)
}
