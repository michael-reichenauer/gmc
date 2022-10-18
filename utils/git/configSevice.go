package git

type configService struct {
	cmd gitCommander
}

func newConfigService(cmd gitCommander) *configService {
	return &configService{cmd: cmd}
}

func (t *configService) ConfigUser(name, email string) error {
	_, err := t.cmd.Git("config", "user.name", name)
	if err != nil {
		return err
	}
	_, err = t.cmd.Git("config", "user.email", email)
	if err != nil {
		return err
	}

	return nil
}
