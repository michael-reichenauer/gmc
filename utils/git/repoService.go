package git

type repoService struct {
	cmd gitCommander
}

func newRepoService(cmd gitCommander) *repoService {
	return &repoService{cmd: cmd}
}

func (t *repoService) InitRepo() error {
	_, err := t.cmd.Git("init", t.cmd.WorkingDir())
	return err
}

func (t *repoService) InitRepoBare() error {
	_, err := t.cmd.Git("init", "--bare", t.cmd.WorkingDir())
	return err
}
