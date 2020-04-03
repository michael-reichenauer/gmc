package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"strings"
)

type MessageBox struct {
	uiHandler      *UI
	messageBoxView *messageBoxView
	text           string
}

func NewMessageBox(uiHandler *UI, text, title string) *MessageBox {
	return &MessageBox{
		uiHandler:      uiHandler,
		messageBoxView: newMessageBoxView(uiHandler, title, true),
		text:           text}
}

func (h *MessageBox) Show() {
	text := h.text + "\n\n\n[OK]"
	lines := strings.Split(text, "\n")

	bounds, err := h.getBounds(lines)
	if err != nil {
		log.Warnf("Failed to show msg box, %v", err)
	}

	h.messageBoxView.show(bounds, lines)
}

func (h *MessageBox) getBounds(lines []string) (Rect, error) {
	windowWidth, windowHeight := h.uiHandler.WindowSize()
	if windowWidth < 4 || windowHeight < 4 {
		return Rect{}, fmt.Errorf("to small window, to shwo message box")
	}
	width := h.maxTextWidth(lines)
	if width < 30 {
		width = 30
	}
	if width > windowWidth-4 {
		width = windowWidth - 4
	}
	height := len(lines)
	if height < 4 {
		height = 4
	}
	if height > windowHeight-4 {
		height = windowHeight - 4
	}
	x := (windowWidth - width) / 2
	y := (windowHeight - height) / 2

	return Rect{x, y, width, height}, nil
}

func (h *MessageBox) maxTextWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}

type messageBoxView struct {
	View
	lines       []string
	ui          *UI
	currentView View
}

func newMessageBoxView(uiHandler *UI, title string, hideCurrent bool) *messageBoxView {
	h := &messageBoxView{ui: uiHandler}
	h.View = uiHandler.NewViewFromPageFunc(h.viewData)
	h.View.Properties().Name = "AboutView"
	h.View.Properties().HasFrame = true
	h.View.Properties().Title = title
	h.View.Properties().HideCurrent = hideCurrent
	return h
}

func (h *messageBoxView) show(bounds Rect, lines []string) {
	h.lines = lines
	h.currentView = h.ui.CurrentView()
	h.SetKey(gocui.KeyEsc, gocui.ModNone, h.onClose)
	h.SetKey(gocui.KeyEnter, gocui.ModNone, h.onEnter)
	h.Show(bounds)
	h.SetCurrentView()
	h.NotifyChanged()
}

func (h *messageBoxView) viewData(viewPort ViewPage) ViewText {
	var lines []string
	length := viewPort.FirstLine + viewPort.Height
	if length > len(h.lines) {
		length = len(h.lines)
	}

	for i := viewPort.FirstLine; i < length; i++ {
		lines = append(lines, utils.Text(h.lines[i], viewPort.Width))
	}
	return ViewText{Lines: lines, Total: len(h.lines)}
}

func (h *messageBoxView) onClose() {
	h.Close()
	h.ui.SetCurrentView(h.currentView)
}

func (h *messageBoxView) onEnter() {
	h.onClose()
}
