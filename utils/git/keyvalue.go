package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type keyValueService struct {
	cmd           gitCommander
	remoteService *remoteService
}

func newKeyValue(cmd gitCommander, remoteService *remoteService) *keyValueService {
	return &keyValueService{cmd: cmd, remoteService: remoteService}
}

func (t *keyValueService) getValue(key string) (string, error) {

	value, err := t.cmd.Git("cat-file", "-p", t.getKeyPath(key))
	if err != nil {
		return "", err
	}
	return value, nil
}

func (t *keyValueService) pushValue(key string) error {
	return t.remoteService.pushRefForce(t.getKeyPath(key))
}

func (t *keyValueService) pullValue(key string) error {
	return t.remoteService.pullRef(t.getKeyPath(key))
}

func (t *keyValueService) setValue(key, value string) error {
	// Store value as a temp file in the git repo
	tmpFile, err := t.createTmpFile()
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	tmpPath := tmpFile.Name()

	err = ioutil.WriteFile(tmpPath, []byte(value), 0644)
	if err != nil {
		return err
	}

	// Store the temp file in the git database (returns an object id)
	objectId, err := t.cmd.Git("hash-object", "-w", tmpPath)
	if err != nil {
		return err
	}
	objectId = strings.TrimSpace(objectId)

	// Add a ref pointer to the stored object for easier retrieval
	_, err = t.cmd.Git("update-ref", t.getKeyPath(key), objectId)
	if err != nil {
		return err
	}

	return nil
}

func (t *keyValueService) getKeyPath(key string) string {
	return fmt.Sprintf("refs/gmc-metadata-key-value/%s", key)
}

func (t *keyValueService) createTmpFile() (f *os.File, err error) {
	gitRepoPath := filepath.Join(t.cmd.WorkingDir(), ".git")
	return ioutil.TempFile(gitRepoPath, "gmc-tmp-key-value-")
}
