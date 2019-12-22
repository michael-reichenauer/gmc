package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"testing"
)

type mock struct {
	notified func()
}

func (m *mock) NotifyChanged() {
	m.notified()
}
func TestView(t *testing.T) {
	m := viewmodel.NewModel(`C:\Work Files\GitMind`)
	var vm *repoVM
	vm = newRepoVM(m, &mock{func() {
		vd, _ := vm.GetRepoPage(100, 0, 20, 0)
		fmt.Printf(vd.text)
	}})
	vm.Load()

}

func TestViewCurrent(t *testing.T) {
	m := viewmodel.NewModel(utils.CurrentDir())
	var vm *repoVM
	done := make(chan interface{})
	vm = newRepoVM(m, &mock{func() {
		vd, _ := vm.GetRepoPage(100, 0, 20, 0)
		fmt.Printf(vd.text)
		close(done)
	}})
	vm.Load()
	<-done
}
