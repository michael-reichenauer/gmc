package git

import (
	"encoding/json"
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGit_GetRepo(t *testing.T) {
	tests.ManualTest(t)

	cmd := newRecorderCmd(newGitCmd(utils.CurrentDir()))
	gitService := NewGitWith(cmd)
	_, err := gitService.GetRepo()
	assert.NoError(t, err)

	gs := NewGitWith(newMockCmd(cmd.String()))
	repo, err := gs.GetRepo()
	assert.NoError(t, err)
	t.Logf("%+v", repo)
}

type mockCmd struct {
	responses responses
}

func newMockCmd(text string) *mockCmd {
	var responses responses
	if err := json.Unmarshal([]byte(text), &responses); err != nil {
		panic(log.Fatal(err))
	}

	return &mockCmd{responses: responses}
}

func (t *mockCmd) Git(args ...string) (string, error) {
	rsp, ok := t.responses.Cmds[fmt.Sprintf("%v", args)]
	if !ok {
		return "", fmt.Errorf("no command output for: %v", args)
	}
	var err error
	if rsp.Error != "" {
		err = fmt.Errorf(rsp.Error)
	}
	return rsp.Output, err
}

func (t *mockCmd) RepoPath() string {
	return t.responses.Path
}

func (t *mockCmd) ReadFile(path string) (string, error) {
	rsp, ok := t.responses.Cmds[path]
	if !ok {
		return "", fmt.Errorf("no file: %v", path)
	}
	var err error
	if rsp.Error != "" {
		err = fmt.Errorf(rsp.Error)
	}
	return rsp.Output, err
}

type resp struct {
	Output string
	Error  string
}

type responses struct {
	Path string
	Cmds map[string]resp
}

type recorderCmd struct {
	cmd       gitCommander
	responses responses
}

func newRecorderCmd(cmd gitCommander) *recorderCmd {
	return &recorderCmd{cmd: cmd, responses: responses{Path: cmd.RepoPath(), Cmds: make(map[string]resp)}}
}

func (t *recorderCmd) Git(args ...string) (string, error) {
	output, err := t.cmd.Git(args...)
	e := ""
	if err != nil {
		e = err.Error()
	}
	t.responses.Cmds[fmt.Sprintf("%v", args)] = resp{Output: output, Error: e}

	return output, err
}

func (t *recorderCmd) RepoPath() string {
	t.responses.Path = t.cmd.RepoPath()
	return t.responses.Path
}

func (t *recorderCmd) ReadFile(path string) (string, error) {
	output, err := t.cmd.ReadFile(path)
	e := ""
	if err != nil {
		e = err.Error()
	}
	t.responses.Cmds[path] = resp{Output: output, Error: e}
	return output, err
}

func (t *recorderCmd) String() string {
	text, _ := json.Marshal(t.responses)
	return string(text)
}
