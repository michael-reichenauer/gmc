package repoview

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/repoview/viewmodel"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/gitlib"
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
// func TestViewCurrent(t *testing.T) {
// 	gitlib.EnableTracing("")
// 	m := viewmodel.NewModel(utils.CurrentDir())
// 	var vm *repoVM
// 	done := make(chan interface{})
// 	vm = newRepoVM(m, &mock{func() {
// 		vd, _ := vm.GetRepoPage(ui.ViewPort{FirstLine: 0, Height: 20, CurrentIndex: 1, Width: 120})
// 		fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
// 		//for _, l := range vd.lines {
// 		//	fmt.Printf("length: %d\n", strings.Count(l, ""))
// 		//}
// 		close(done)
// 	}})
// 	vm.Load()
// 	<-done
// }

func TestViewTraced(t *testing.T) {
	gitlib.EnableReplay("")
	var trace trace

	traceBytes := utils.MustFileRead(filepath.Join(gitlib.TracePath(""), "repovm"))
	utils.MustJsonUnmarshal(traceBytes, &trace)

	m := viewmodel.NewModel(trace.RepoPath)
	vm := newRepoVM(m, nil)
	vm.LoadWithBranches(trace.BranchNames)
	vd, _ := vm.GetRepoPage(trace.ViewPage)
	fmt.Printf("%s\n", strings.Join(vd.lines, "\n"))
}
