package cui

var menuSeparator = MenuItem{isSeparator: true}

func MenuSeparator(text string) MenuItem {
	return MenuItem{isSeparator: true, Text: text}

}

type MenuItem struct {
	Text        string
	Title       string
	Key         string
	Action      func()
	Items       []MenuItem
	ItemsFunc   func() []MenuItem
	isSeparator bool
	ReuseBounds bool
}

type Menu interface {
	Add(item ...MenuItem)
	AddItems(items []MenuItem)
	Show(x int, y int)
	OnClose(onClose func())
}

type menu struct {
	menuView *menuView
}

func newMenu(uiHandler *ui, title string) *menu {
	return &menu{menuView: newMenuView(uiHandler, title, nil)}
}

func (h *menu) Add(items ...MenuItem) {
	h.menuView.addItems(items)
}

func (h *menu) AddItems(items []MenuItem) {
	h.menuView.addItems(items)
}

func (h *menu) Show(x, y int) {
	h.menuView.show(Rect{X: x, Y: y})
}

func (h *menu) OnClose(onClose func()) {
	h.menuView.onClose = onClose
}
