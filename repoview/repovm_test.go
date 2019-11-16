package repoview

import (
	"fmt"
	"gmc/utils"
	"testing"
)

func TestView(t *testing.T) {
	vm := newRepoVM(`C:\Work Files\GitMind`)
	vm.Load()
	vd, _ := vm.GetRepoPage(100, 0, 20, 0)
	fmt.Printf(vd.text)
}

func TestViewCurrent(t *testing.T) {
	vm := newRepoVM(utils.CurrentDir())
	vm.Load()
	vd, _ := vm.GetRepoPage(100, 0, 18, 1)
	fmt.Printf(vd.text)
}
