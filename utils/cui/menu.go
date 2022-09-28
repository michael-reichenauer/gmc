package cui

var SeparatorMenuItem = MenuItem{isSeparator: true}

type MenuItem struct {
	Text         string
	Title        string
	Key          string
	Action       func()
	SubItems     []MenuItem
	SubItemsFunc func() []MenuItem
	isSeparator  bool
	ReuseBounds  bool
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
