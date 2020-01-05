package installation

import (
	"github.com/michael-reichenauer/gmc/common/config"
	"testing"
)

func TestAutoUpdate_Check(t *testing.T) {
	c := config.NewConfig()
	au := NewAutoUpdate(c, "v0.3")
	au.CheckRemoteReleases()
}

func TestAutoUpdate_UpdateIfNewer(t *testing.T) {
	c := config.NewConfig()
	au := NewAutoUpdate(c, "v0.2")
	au.updateIfNewer()
}
