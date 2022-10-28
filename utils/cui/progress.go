package cui

import (
	"fmt"
	"strings"
	"time"

	"github.com/michael-reichenauer/gmc/utils/log"
)

const (
	progressInterval = 500 * time.Millisecond
	waitMark         = " ● ● ● ● "
	waitMark2        = " ●"
	showIconTimeout  = 500 * time.Millisecond
	showFullTimeout  = 15 * time.Second
	line1            = "┌───────────────────┐"
	line2            = "└───────────────────┘"
)

var (
	instance          *progress
	instanceStartTime time.Time
	instanceCloseTime time.Time
)

type Progress interface {
	Close()
}

type progress struct {
	ui           *ui
	text         string
	view         View
	showCount    int
	length       int
	startTime    time.Time
	showProgress bool
}

func showProgress(ui *ui, format string, v ...interface{}) Progress {
	text := fmt.Sprintf(format, v...)

	if instance == nil {
		startTime := time.Now()
		if time.Since(instanceCloseTime) < showIconTimeout {
			startTime = instanceStartTime
		}
		instanceStartTime = startTime
		instance = newProgress(ui, startTime)
	}
	instance.showCount++
	log.Debugf("Start progress for %q, #%d ...", text, instance.showCount)

	if instance.showCount == 1 {
		instance.show()
	}
	instance.SetText(text)
	instance.view.SetTop()
	return instance
}

func newProgress(ui *ui, startTime time.Time) *progress {
	t := &progress{ui: ui, length: 0, startTime: startTime}
	t.view = t.newView()
	return t
}

func (t *progress) show() {
	log.Debugf("Show progress %q at %v", t.text, t.startTime)
	t.view.Show(CenterBounds(30, 3, 30, 3))
	t.view.SetTop()
	t.view.SetCurrentView()
	t.updateProgress()
}

func (t *progress) newView() View {
	view := t.ui.NewViewFromTextFunc(t.textFunc)
	view.Properties().HasFrame = false
	view.Properties().Name = "Progress"
	view.Properties().HideHorizontalScrollbar = true
	view.Properties().HideVerticalScrollbar = true
	view.Properties().HideCurrentLineMarker = true
	return view
}

func (t *progress) SetText(text string) {
	log.Debugf("Progress text: %q", text)
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
	log.Debugf("End progress #%d", instance.showCount)
	instance.showCount--
	if instance.showCount == 0 {
		instanceCloseTime = time.Now()
		log.Debugf("Close Progress %v", time.Since(t.startTime))
		t.view.Close()
		t.view = nil
		instance = nil
	}
}

func (t *progress) textFunc(ViewPage) string {
	sinceStart := time.Since(t.startTime)
	if sinceStart < showIconTimeout {
		// Show no progress for a show while in case operation completes fast
		return ""
	}
	t.view.SetTop()

	//if sinceStart < showFullTimeout {
	length := t.length - 2
	if length < 0 {
		length = 0
	}

	mark := fmt.Sprintf("%s\n│%-19s│\n%s", line1, strings.Repeat(waitMark2, 1+length%9), line2)
	return MagentaDk(mark)
	//}

	// // log.Infof("Show full progress for %q, ...", t.text)
	// if !t.showProgress {
	// 	t.showProgress = true
	// 	t.length = 0
	// }

	// // Show full progress
	// t.view.ShowFrame(true)
	// pt := strings.Repeat("━", t.length)
	// return fmt.Sprintf("%s\n%s", t.text, MagentaDk(pt))
}

func (t *progress) updateProgress() {
	t.ui.Post(func() {
		if t.view == nil {
			return
		}
		// Calculate length of progress bar
		p := t.view.ViewPage()
		t.length = (t.length + 1) % (p.Width + 2)

		t.view.NotifyChanged()
		time.AfterFunc(progressInterval, t.updateProgress)
	})
}
