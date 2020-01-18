package ui

type Item struct {
	Text     string
	Key      string
	Action   func()
	SubItems []Item
}

type Menu struct {
	menuView *menuView
}

func NewMenu(uiHandler *UI, items []Item) *Menu {
	return &Menu{menuView: newMenuView(uiHandler, nil, items, 5, 5)}
}

func (h *Menu) Show() {
	h.menuView.show()
}
