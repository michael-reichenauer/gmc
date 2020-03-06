package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"strings"
	"testing"
)

type mock struct {
	uiWork chan func()
}

func (m *mock) NotifyChanged() {
	close(m.uiWork)
}
func (m *mock) PostOnUIThread(f func()) {
	select {
	case m.uiWork <- f:
	default:
	}
}

func TestViewCurrent(t *testing.T) {
	cs := config.NewConfig("0.0", "")
	m := &mock{uiWork: make(chan func())}
	vm := newRepoVM(m, nil, cs, "")
	vm.load()
	vm.refresh()
	for f := range m.uiWork {
		f()
	}
	vd, _ := vm.GetRepoPage(ui.ViewPage{Height: 20, Width: 120})
	fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
}

//
// func TestView(t *testing.T) {
// 	cs := config.NewConfig()
// 	m := viewmodel.NewService(cs, `C:\code\gmc2`)
// 	vm := newRepoVM(m, nil, "")
// 	vm.LoadWithBranches([]string{})
// 	vd, _ := vm.GetRepoPage(ui.ViewPage{Height: 20, Width: 120})
// 	fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
// }
//
// func TestSavedState(t *testing.T) {
// 	cs := config.NewConfig()
// 	git.EnableReplay("")
// 	var trace trace
//
// 	traceBytes := utils.MustFileRead(filepath.Join(git.TracePath(""), "repovm"))
// 	utils.MustJsonUnmarshal(traceBytes, &trace)
//
// 	m := viewmodel.NewService(cs, trace.RepoPath)
// 	vm := newRepoVM(m, nil, "")
// 	vm.LoadWithBranches(trace.BranchNames)
// 	vd, _ := vm.GetRepoPage(trace.ViewPage)
// 	fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
// }
