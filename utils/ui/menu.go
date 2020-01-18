package ui

type MenuItem struct {
	Text     string
	Key      string
	Action   func()
	SubItems []MenuItem
}

type Menu struct {
	menuView *menuView
}

func NewMenu(uiHandler *UI, items []MenuItem) *Menu {
	return &Menu{menuView: newMenuView(uiHandler, nil, items, 5, 5)}
}

func (h *Menu) Show() {
	h.menuView.show()
}
