package gitlib

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStatusFromCurrentDir(t *testing.T) {
	status, err := newStatus(newGitCmd(utils.CurrentDir())).getStatus()
	assert.NoError(t, err)

	t.Log(status)
}
