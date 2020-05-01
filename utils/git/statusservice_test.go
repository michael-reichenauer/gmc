package git

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStatusFromCurrentDir(t *testing.T) {
	tests.ManualTest(t)
	status, err := newStatus(newGitCmd(utils.CurrentDir())).getStatus()
	assert.NoError(t, err)

	t.Log(status)
}
