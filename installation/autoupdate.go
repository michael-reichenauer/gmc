package installation

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/telemetry"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	releasesUri    = "https://api.github.com/repos/michael-reichenauer/gmc/releases"
	binNameWindows = "gmc.exe"
	binNameLinux   = "gmc_linux"
	binNameMac     = "gmc_mac"
	tmpSuffix      = ".tmp"
)

// Type used when parsing latest version information json
type Release struct {
	Tag_name     string
	Draft        bool
	Prerelease   bool
	Published_at string
	Body         string
	Assets       []Asset
}

// Type used when parsing latest version information json
type Asset struct {
	Name                 string
	Download_count       int
	Browser_download_url string
}

type autoUpdate struct {
	config  *config.Service
	tel     *telemetry.Telemetry
	version string
}

func NewAutoUpdate(config *config.Service, tel *telemetry.Telemetry, version string) *autoUpdate {
	return &autoUpdate{config: config, tel: tel, version: version}
}

func (h *autoUpdate) Start() {
	if runtime.GOOS != "windows" {
		// Only support for windows
		log.Infof("Auto update only supported on windows")
		h.tel.SendEvent("autoupdate-os-not-supported")
		return
	}
	h.cleanTmpFiles()
	go h.periodicCheckForUpdatesRoutine()
}

func (h *autoUpdate) periodicCheckForUpdatesRoutine() {
	h.tel.SendEvent("autoupdate-start-periodic-check")
	for {
		h.UpdateIfAvailable()
		time.Sleep(5 * time.Minute)
	}
}

func (h *autoUpdate) UpdateIfAvailable() {
	conf := h.config.GetConfig()
	if conf.DisableAutoUpdate {
		log.Infof("Auto update disabled")
		h.tel.SendEvent("autoupdate-disabled")
		return
	}
	log.Infof("Check updates for %s, allow preview=%v ...", h.version, conf.AllowPreview)
	h.tel.SendEventf("autoupdate-check", "%s, allow preview=%v", h.version, conf.AllowPreview)

	h.checkRemoteReleases()

	if !h.isUpdateAvailable(conf.AllowPreview) {
		h.tel.SendEvent("autoupdate-no-update")
		return
	}

	h.tel.SendEvent("autoupdate-update-available")
	h.update(conf.AllowPreview)
}

func (h *autoUpdate) isUpdateAvailable(allowPreview bool) bool {
	state := h.config.GetState()
	release := h.selectRelease(state, allowPreview)
	if release.Version == "" {
		log.Infof("No remote release available")
		return false
	}
	if len(release.Assets) == 0 {
		log.Warnf("No binaries for %s", release.Version)
		return false
	}

	if !h.isNewer(release.Version, h.version) {
		log.Infof("No update available, local %s>=%s remote, allow preview=%v",
			h.version, release.Version, allowPreview)
		return false
	}
	log.Infof("Update available, local %s<%s remote (preview=%v)", h.version, release.Version, release.Preview)
	return true
}

func (h *autoUpdate) update(allowPreview bool) {
	state := h.config.GetState()
	release := h.selectRelease(state, allowPreview)

	if state.InstalledVersion == release.Version {
		// Already downloaded and installed the newer version
		log.Infof("Already downloaded and installed the remote version %s", release.Version)
		return
	}

	downloadPath, err := h.download(release)
	if err != nil {
		log.Warnf("Failed to download %s for %s, %v", release.Version, runtime.GOOS, err)
		return
	}
	h.tel.SendEvent("autoupdate-downloaded")

	h.replaceRunningBinary(release, downloadPath)
}

func (h *autoUpdate) selectRelease(state config.State, allowPreview bool) config.Release {
	release := state.StableRelease
	if allowPreview &&
		len(state.PreRelease.Assets) > 0 &&
		h.isNewer(state.StableRelease.Version, state.PreRelease.Version) {
		// user allow preview versions, and the preview version is newer
		release = state.PreRelease
	}
	return release
}

func (h *autoUpdate) checkRemoteReleases() {
	state := h.config.GetState()

	body, etag, err := h.httpGet(releasesUri, state.ReleasesEtag)
	if err != nil {
		log.Warnf("Failed to get release info %s, %v", releasesUri, err)
		return
	}
	if len(body) == 0 {
		return
	}

	log.Infof("Got new release info from %s", releasesUri)
	// Parse release info
	var releases []Release
	err = json.Unmarshal(body, &releases)
	if err != nil {
		log.Warnf("Failed to parse release info, %v", err)
		return
	}

	// Cache release info (and the corresponding ETag)
	preRelease, _ := h.getPreRelease(releases)
	stableRelease, _ := h.getStableRelease(releases)
	log.Infof("Pre-release info %q, %d files", preRelease.Tag_name, len(preRelease.Assets))
	log.Infof("Stable-release info %q, %d files", stableRelease.Tag_name, len(stableRelease.Assets))

	h.config.SetState(func(s *config.State) {
		s.ReleasesEtag = etag
		s.PreRelease = h.toConfigRelease(preRelease)
		s.StableRelease = h.toConfigRelease(stableRelease)
	})
}

func (h *autoUpdate) getPreRelease(releases []Release) (Release, bool) {
	for _, r := range releases {
		if r.Prerelease {
			return r, true
		}
	}
	return Release{}, false
}

func (h *autoUpdate) getStableRelease(releases []Release) (Release, bool) {
	for _, r := range releases {
		if !r.Prerelease {
			return r, true
		}
	}
	return Release{}, false
}

func (h *autoUpdate) toConfigRelease(release Release) config.Release {
	log.Infof("Release info: %+v", release)
	var assets []config.Asset
	for _, a := range release.Assets {
		log.Infof("Release %s, %s, downloaded: %d:", release.Tag_name, a.Name, a.Download_count)
		assets = append(assets, config.Asset{
			Name: a.Name,
			Url:  a.Browser_download_url,
		})
	}
	return config.Release{
		Version: release.Tag_name,
		Preview: release.Prerelease,
		Assets:  assets,
	}
}

func (h *autoUpdate) isNewer(v1 string, v2 string) bool {
	if strings.HasPrefix(v1, "v") {
		v1 = v1[1:]
	}
	if strings.HasPrefix(v2, "v") {
		v2 = v2[1:]
	}
	if v1 != "" && v2 == "" {
		return false
	}
	if v1 == "" && v2 != "" {
		return true
	}
	version1, err := version.NewVersion(v1)
	if err != nil {
		return false
	}
	version2, err := version.NewVersion(v2)
	if err != nil {
		return false
	}
	return version1.GreaterThan(version2)
}

func (h *autoUpdate) replaceRunningBinary(release config.Release, downloadedPath string) {
	// Switch current binary
	// Move current binary to temp path
	tmpPath := h.tempBinPath()
	currentPath := utils.BinPath()
	log.Infof("Moving %q to %q ...", currentPath, tmpPath)
	err := os.Rename(currentPath, tmpPath)
	if err != nil {
		log.Warnf("Failed to move running binary to %sq %v", tmpPath, err)
		return
	}
	log.Infof("Moved %q to %q", currentPath, tmpPath)

	// Move downloaded binary to current binary path
	log.Infof("Moving %q to %q ...", downloadedPath, currentPath)
	err = os.Rename(downloadedPath, currentPath)
	if err != nil {
		log.Warnf("Failed to move downloaded binary to %q, %v", currentPath, err)
		return
	}
	log.Infof("Moved %s to %s", downloadedPath, currentPath)

	// Store info that downloaded binary has been installed to prevent repeating for this version
	h.config.SetState(func(s *config.State) {
		s.InstalledVersion = release.Version
	})
	log.Infof("Replaced %s->%s at %q", h.version, release.Version, currentPath)
	h.tel.SendEvent("autoupdate-replaced")
}

func (h *autoUpdate) download(release config.Release) (string, error) {
	log.Infof("Downloading %s ...", release.Version)
	downloadPath := h.versionedBinPath(release.Version)
	if utils.FileExists(downloadPath) {
		// File already downloaded
		log.Infof("File already downloaded: %q", downloadPath)
		return downloadPath, nil
	}

	// Get the binary uri for the current os type
	binURI, err := h.getBinURI(release)
	if err != nil {
		return "", err
	}

	// Download the binary
	binary, _, err := h.httpGet(binURI, "")
	if err != nil {
		return "", fmt.Errorf("failed to download %q,%v", binURI, err)
	}
	utils.MustFileWrite(downloadPath, binary)

	// Make the binary executable (not needed on windows)
	err = h.makeBinaryExecutable(downloadPath)
	if err != nil {
		return "", err
	}
	log.Infof("Downloaded %q", downloadPath)
	return downloadPath, nil
}

func (h *autoUpdate) getBinURI(release config.Release) (string, error) {
	// Determine the binary based on the os
	var binaryName string
	switch runtime.GOOS {
	case "windows":
		binaryName = binNameWindows
	case "linux":
		binaryName = binNameLinux
	case "darwin":
		binaryName = binNameMac
	default:
		return "", fmt.Errorf("unsupported os %q", runtime.GOOS)
	}

	// Get the binary uri for the specified binary
	var binURI string
	for _, a := range release.Assets {
		if a.Name == binaryName {
			binURI = a.Url
			break
		}
	}
	if binURI == "" {
		// No binary available
		return "", fmt.Errorf("no binary uri for %s %s", binaryName, release.Version)
	}
	return binURI, nil
}

func (h *autoUpdate) versionedBinPath(version string) string {
	return utils.BinPath() + "." + version + tmpSuffix
}

func (h *autoUpdate) tempBinPath() string {
	return utils.BinPath() + "." + utils.RandomString(10) + tmpSuffix
}

func (h *autoUpdate) cleanTmpFiles() {
	binPath := utils.BinPath()
	binDir := filepath.Dir(binPath)
	var tmpFiles []string
	err := filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(path, binPath) && filepath.Ext(path) == tmpSuffix {
			tmpFiles = append(tmpFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Warnf("Failed to get temp files in bin dir")
		return
	}
	for _, tempFile := range tmpFiles {
		log.Infof("Removing tmp file: %q", tempFile)
		var err = os.Remove(tempFile)
		if err != nil {
			log.Infof("Failed to remove %q, %s", tempFile, err)
		}
	}
}

func (h *autoUpdate) httpGet(url, requestEtag string) (bytes []byte, etag string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(log.Fatal(err))
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("user-agent", "gmc")
	if requestEtag != "" {
		req.Header.Add("If-None-Match", requestEtag)
	}

	client := &http.Client{Transport: &http.Transport{Proxy: nil}}
	resp, err := client.Do(req)
	if err != nil {
		log.Infof("No contact with %s, %s", url, err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	etagValue := resp.Header["Etag"]
	if len(etagValue) > 0 {
		etag = etagValue[0]
		if strings.HasPrefix(etag, "W/") {
			etag = etag[2:]
		}
	}

	if resp.StatusCode == http.StatusNotModified {
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Infof("Invalid status for %s: %d", url, resp.StatusCode)
		return
	}

	bytes, err = ioutil.ReadAll(resp.Body)
	return
}

func (h *autoUpdate) makeBinaryExecutable(path string) error {
	if runtime.GOOS == "windows" {
		// Not needed on windows
		return nil
	}
	err := os.Chmod(path, 0700)
	if err != nil {
		log.Warnf("Failed to make %q binary executable, %v", path, err)
		return err
	}
	return nil
}
