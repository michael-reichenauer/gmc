package ui

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/timer"
	"strings"
	"time"
)

const progressInterval = 500 * time.Millisecond

type Progress interface {
	Close()
}
type progress struct {
	ui        *ui
	text      string
	view      View
	length    int
	startTime *timer.Timer
}

func newProgress(ui *ui) *progress {
	t := &progress{ui: ui, length: 1, startTime: timer.Start()}
	t.view = t.newView()
	return t
}

func (t *progress) show() {
	log.Infof("Show progress %q", t.text)
	t.view.Show(CenterBounds(30, 3, 30, 3))
	t.view.SetTop()
	t.view.SetCurrentView()
	time.AfterFunc(progressInterval, t.elapsed)
}

func (t *progress) newView() View {
	view := t.ui.NewViewFromTextFunc(t.textFunc)
	view.Properties().HasFrame = true
	view.Properties().Name = "Progress"
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (t *progress) SetText(text string) {
	log.Infof("Progress text: %q", text)
	// Calculate margin between text and progress indicator (max two lines of text)
	lines := strings.Split(text, "\n")
	if len(lines) > 2 {
		lines = lines[:2]
	}
	margin := strings.Repeat("\n", 2-len(lines))

	t.text = text + margin
	t.view.NotifyChanged()
}

func (t *progress) Close() {
	log.Infof("Close Progress %s", t.startTime)
	t.view.Close()
	t.view = nil
}

func (t *progress) textFunc(ViewPage) string {
	pt := strings.Repeat("━", t.length)
	return fmt.Sprintf("%s\n%s", t.text, MagentaDk(pt))
}

func (t *progress) elapsed() {
	t.ui.PostOnUIThread(func() {
		if t.view == nil {
			return
		}
		// Calculate length of progress bar
		p := t.view.ViewPage()
		t.length = (t.length + 1) % (p.Width + 2)

		t.view.NotifyChanged()
		time.AfterFunc(progressInterval, t.elapsed)
	})
}
