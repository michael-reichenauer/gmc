package git

import (
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStatusFromCurrentDir(t *testing.T) {
	status, err := getStatus(utils.CurrentDir())
	assert.NoError(t, err)

	t.Log(status)
}
