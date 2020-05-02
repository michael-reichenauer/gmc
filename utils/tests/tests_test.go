package tests

import (
	"testing"
)

func TestGetTempFolder(t *testing.T) {
	path := GetTempFolder()
	t.Log(path)
}
