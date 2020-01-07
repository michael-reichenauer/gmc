package viewmodel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	model := NewModel(`C:\Work Files\GitMind`)
	model.Load()
	vp, err := model.GetRepoViewPort(0, 18, 0)
	assert.NoError(t, err)
	assert.NotEqual(t, len(vp.Commits), 0)
	for _, c := range vp.Commits {
		t.Logf("%s", c.String())
	}
}
