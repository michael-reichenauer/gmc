package program

import (
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/repoview"
	"github.com/michael-reichenauer/gmc/repoview/viewrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/ui"
)

type MainWindow struct {
	ui            ui.UI
	configService *config.Service
	model         *viewrepo.Service
	commitView    *repoview.CommitView
}

func NewMainWindow(ui ui.UI, configService *config.Service) *MainWindow {
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

func (h *MainWindow) MainMenuItem() ui.MenuItem {
	var subItems []ui.MenuItem
	subItems = append(subItems, h.OpenRepoMenuItems()...)
	subItems = append(subItems, ui.MenuItem{Text: "About", Action: h.showAbout})
	menuItem := ui.MenuItem{
		Text:     "Main Menu",
		Title:    "Main Menu",
		SubItems: subItems,
	}

	return menuItem
}

func (h *MainWindow) showAbout() {
	h.ui.ShowMessageBox("About", "gmc version %s\n%s", h.configService.ProgramVersion, git.Version())
}
