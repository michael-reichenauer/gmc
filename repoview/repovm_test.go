package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"path/filepath"
	"strings"
	"testing"
)

type mock struct {
	notified func()
}

func (m *mock) NotifyChanged() {
	m.notified()
}

//
// func TestView(t *testing.T) {
// 	m := viewmodel.NewModel(`C:\Work Files\GitMind`)
// 	var vm *repoVM
// 	done := make(chan interface{})
// 	vm = newRepoVM(m, &mock{func() {
// 		vd, _ := vm.GetRepoPage(ui.ViewPort{FirstLine: 0, Height: 20, CurrentIndex: 1, Width: 120})
// 		fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
// 		close(done)
// 	}})
// 	vm.Load()
// 	<-done
// }
//
func TestViewCurrent(t *testing.T) {
	cs := config.NewConfig()
	m := viewmodel.NewModel(cs, utils.CurrentDir())
	vm := newRepoVM(m, nil)
	vm.LoadWithBranches([]string{})
	vd, _ := vm.GetRepoPage(ui.ViewPage{Height: 20, Width: 120})
	fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
}

func TestView(t *testing.T) {
	cs := config.NewConfig()
	m := viewmodel.NewModel(cs, `C:\code\gmc2`)
	vm := newRepoVM(m, nil)
	vm.LoadWithBranches([]string{})
	vd, _ := vm.GetRepoPage(ui.ViewPage{Height: 20, Width: 120})
	fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
}

func TestSavedState(t *testing.T) {
	cs := config.NewConfig()
	git.EnableReplay("")
	var trace trace

	traceBytes := utils.MustFileRead(filepath.Join(git.TracePath(""), "repovm"))
	utils.MustJsonUnmarshal(traceBytes, &trace)

	m := viewmodel.NewModel(cs, trace.RepoPath)
	vm := newRepoVM(m, nil)
	vm.LoadWithBranches(trace.BranchNames)
	vd, _ := vm.GetRepoPage(trace.ViewPage)
	fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
}
