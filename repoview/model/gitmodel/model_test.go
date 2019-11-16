package gitmodel

import "testing"

func Test(t *testing.T) {
	model := NewModel(`C:\Work Files\GitMind`)
	model.Load()
	repo := model.GetRepo()
	for _, c := range repo.Commits[:20] {
		t.Log(c)
	}
}
