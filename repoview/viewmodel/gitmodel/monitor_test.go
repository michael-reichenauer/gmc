package gitmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// test.d vv
func TestMonitor(t *testing.T) {
	current := utils.CurrentDir()
	rootPath, err := git.WorkingFolderRoot(current)
	assert.NoError(t, err)
	mon := newMonitor(rootPath)
	assert.NoError(t, mon.Start())
	for e := range mon.StatusChange {
		fmt.Printf("Event: %d\n", e)
		time.Sleep(100 * time.Millisecond)
	}
	mon.Close()
}