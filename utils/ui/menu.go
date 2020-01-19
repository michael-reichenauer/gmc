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

func NewMenu(uiHandler *UI, title string) *Menu {
	return &Menu{menuView: newMenuView(uiHandler, title, nil)}
}

func (h *Menu) Add(items ...MenuItem) {
	h.menuView.addItems(items)
}

func (h *Menu) AddItems(items []MenuItem) {
	h.menuView.addItems(items)
}

func (h *Menu) Show(x, y int) {
	h.menuView.show(x, y)
}
