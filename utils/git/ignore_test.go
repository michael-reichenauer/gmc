package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIgnoreHandler_IsIgnored(t *testing.T) {
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
		//fmt.Printf("visited file or dir: %q\n", path)
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %v\n", err)
		return
	}
}
