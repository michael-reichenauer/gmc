package ui

type Item struct {
	text     string
	key      string
	action   func()
	subItems []Item
}

type Menu struct {
	uiHandler       *UI
	currentViewName string
	menuView        *menuView
}

func NewMenu(uiHandler *UI) *Menu {
	h := &Menu{
		uiHandler: uiHandler,
		menuView:  newMenuView(uiHandler, nil, 5, 5),
	}
	return h
}

func (h *Menu) Show() {
	h.menuView.show()
}
