package tests

import (
	"testing"
)

func TestGetTempFolder(t *testing.T) {
	path := CreateTempFolder()
	defer CleanTemp()

	t.Log(path)
}
