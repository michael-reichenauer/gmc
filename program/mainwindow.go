package program

import (
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/cviews"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/viewrepo"
)

type MainWindow struct {
	ui            cui.UI
	configService *config.Service
	model         *viewrepo.Service
	commitView    *cviews.CommitView
}

func NewMainWindow(ui cui.UI, configService *config.Service) *MainWindow {
	return &MainWindow{
		ui:            ui,
		configService: configService,
	}
}

func (h *MainWindow) Show() {
	workingFolder, err := h.getWorkingFolder()
	if err != nil {
		// Handle error
	}

	h.OpenRepo(workingFolder)
}

func (h *MainWindow) ToggleShowDetails() {

}

func (h *MainWindow) getWorkingFolder() (string, error) {
	folderPath := h.configService.FolderPath
	if folderPath == "" {
		// No specified repo path, use current dir
		folderPath = utils.CurrentDir()
	}
	path, err := git.WorkingFolderRoot(folderPath)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (h *MainWindow) MainMenuItem() cui.MenuItem {
	var subItems []cui.MenuItem
	subItems = append(subItems, h.OpenRepoMenuItems()...)
	subItems = append(subItems, cui.MenuItem{Text: "About", Action: h.showAbout})
	menuItem := cui.MenuItem{
		Text:     "Main Menu",
		Title:    "Main Menu",
		SubItems: subItems,
	}

	return menuItem
}

func (h *MainWindow) showAbout() {
	h.ui.ShowMessageBox("About", "gmc version %s\n%s", h.configService.ProgramVersion, git.Version())
}
