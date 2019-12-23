package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/ui"
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
		vd, _ := vm.GetRepoPage(ui.ViewPort{Lines: 20, Width: 100})
		fmt.Printf(vd.text)
	}})
	vm.Load()

}

func TestViewCurrent(t *testing.T) {
	m := viewmodel.NewModel(utils.CurrentDir())
	var vm *repoVM
	done := make(chan interface{})
	vm = newRepoVM(m, &mock{func() {
		vd, _ := vm.GetRepoPage(ui.ViewPort{Lines: 20, Width: 100})
		fmt.Printf(vd.text)
		close(done)
	}})
	vm.Load()
	<-done
}
