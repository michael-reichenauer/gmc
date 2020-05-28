package cui

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"strings"
	"time"
)

const (
	progressInterval = 500 * time.Millisecond
	waitMark         = " ╭─╮ \n ╰─╯ "
	showIconTimeout  = 1000 * time.Millisecond
	showFullTimeout  = 8 * time.Second
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
	log.Infof("Start progress for %q, ...", text)
	if instance == nil {
		startTime := time.Now()
		if time.Since(instanceCloseTime) < showIconTimeout {
			startTime = instanceStartTime
			log.Infof("Reuse start time at %v", startTime)
		} else {
			log.Warnf("Use new start time %v", startTime)
		}
		instanceStartTime = startTime
		instance = newProgress(ui, startTime)
	}
	instance.showCount++

	if instance.showCount == 1 {
		instance.show()
	}
	instance.view.SetTop()
	instance.SetText(text)
	return instance
}

func newProgress(ui *ui, startTime time.Time) *progress {
	t := &progress{ui: ui, length: 0, startTime: startTime}
	t.view = t.newView()
	return t
}

func (t *progress) show() {
	log.Infof("Show progress %q at %v", t.text, t.startTime)
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
	log.Infof("Stop progress")
	instance.showCount--
	if instance.showCount == 0 {
		instanceCloseTime = time.Now()
		log.Infof("Close Progress %v", time.Since(t.startTime))
		t.view.Close()
		t.view = nil
		instance = nil
	}
}

func (t *progress) textFunc(ViewPage) string {
	sinceStart := time.Since(t.startTime)
	if sinceStart < showIconTimeout {
		// Show no progress for a show while in case operation completes fast
		log.Infof("Show no progress for %q, ...", t.text)
		return ""
	}

	if sinceStart < showFullTimeout {
		// Show just a small wait icon for a while
		log.Infof("Show icon progress for %q, ...", t.text)
		if t.length%2 == 1 {
			return MagentaDk(waitMark)
		} else {
			return Dark(waitMark)
		}
	}

	log.Infof("Show full progress for %q, ...", t.text)
	if !t.showProgress {
		t.showProgress = true
		t.length = 0
	}
	// Show full progress
	t.view.ShowFrame(true)
	pt := strings.Repeat("━", t.length)
	return fmt.Sprintf("%s\n%s", t.text, MagentaDk(pt))
}

func (t *progress) updateProgress() {
	t.ui.PostOnUIThread(func() {
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
