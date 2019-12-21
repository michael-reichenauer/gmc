package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"testing"
)

func TestView(t *testing.T) {
	m := viewmodel.NewModel(`C:\Work Files\GitMind`)
	vm := newRepoVM(m)
	vm.Load()
	vd, _ := vm.GetRepoPage(100, 0, 20, 0)
	fmt.Printf(vd.text)
}

func TestViewCurrent(t *testing.T) {
	m := viewmodel.NewModel(utils.CurrentDir())
	vm := newRepoVM(m)
	vm.Load()
	vd, _ := vm.GetRepoPage(100, 0, 18, 1)
	fmt.Printf(vd.text)
}
