package installation

import (
	"github.com/michael-reichenauer/gmc/common/config"
	"testing"
)

func TestAutoUpdate_Check(t *testing.T) {
	c := config.NewConfig()
	c.Load()
	au := NewAutoUpdate(c)
	au.CheckReleases()
}
