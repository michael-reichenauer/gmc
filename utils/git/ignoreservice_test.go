package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/tests"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIgnoreHandler_IsIgnored(t *testing.T) {
	wf := tests.CreateTempFolder()
	defer tests.CleanTemp()

	igf := wf.Path(".gitignore")
	utils.FileWrite(igf, []byte(`
*.log*
**/*.exe
`))

	utils.FileWrite(wf.Path("a.txt"), []byte("1"))
	utils.FileWrite(wf.Path("b.exe"), []byte("1"))
	utils.FileWrite(wf.Path("c.log"), []byte("1"))

	assert.NoError(t, os.Mkdir(wf.Path("folder"), 0700))
	wf.File("folder", "d.txt").Write("1")
	wf.File("folder", "e.exe").Write("2")
	wf.File("folder", "f.log").Write("3")

	files, _ := utils.ListFilesRecursively(wf.Path())
	assert.Equal(t, 7, len(files))

	ih := newIgnoreHandler(wf.Path())
	var ignored []string
	for _, f := range files {
		if ih.isIgnored(f) {
			ignored = append(ignored, f)
		}
	}
	assert.Equal(t, 3, len(ignored))
	assert.True(t, utils.StringsContains(ignored, wf.Path("b.exe")))
	assert.True(t, utils.StringsContains(ignored, wf.Path("c.log")))
	assert.True(t, utils.StringsContains(ignored, wf.Path("folder", "e.exe")))
}

func TestIgnoreHandler_IsIgnored_Manual(t *testing.T) {
	tests.ManualTest(t)
	ih := newIgnoreHandler(utils.CurrentDir())

	err := filepath.Walk(`C:\code\gmc`, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		path = strings.ReplaceAll(path, "\\", "/")
		if ih.isIgnored(path) {
			fmt.Printf("Ignored %q\n", path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %v\n", err)
		return
	}
}
