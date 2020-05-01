package installation

import (
	"github.com/michael-reichenauer/gmc/utils/tests"
	"testing"
)

func TestAutoUpdate_Check_Manual(t *testing.T) {
	tests.ManualTest(t)
	// c := config.NewConfig()
	// au := NewAutoUpdate(c, "v0.3")
	// au.checkRemoteReleases()
}

func TestAutoUpdate_UpdateIfNewer_Manual(t *testing.T) {
	tests.ManualTest(t)
	// c := config.NewConfig()
	// au := NewAutoUpdate(c, "v0.2")
	// au.updateIfNewer()
}
