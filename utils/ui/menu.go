package ui

type Item struct {
	Text     string
	Key      string
	Action   func()
	SubItems []Item
}

type Menu struct {
	uiHandler       *UI
	currentViewName string
	menuView        *menuView
}

func NewMenu(uiHandler *UI, items []Item) *Menu {
	h := &Menu{
		uiHandler: uiHandler,
		menuView:  newMenuView(uiHandler, items, 5, 5),
	}
	return h
}

func (h *Menu) Show() {
	h.menuView.show()
}
