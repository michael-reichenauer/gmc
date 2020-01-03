package installation

import (
	"encoding/json"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	//latestUri   = "https://api.github.com/repos/michael-reichenauer/GitMind/releases/latest"
	releasesUri = "https://api.github.com/repos/michael-reichenauer/GitMind/releases"
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
	config *config.Service
}

func NewAutoUpdate(config *config.Service) *autoUpdate {
	return &autoUpdate{config: config}
}

func (h *autoUpdate) CheckReleases() {
	configInfo := h.config.Get()

	body, etag, err := h.httpGet(releasesUri, configInfo.ReleasesEtag)
	if err != nil {
		log.Warnf("Failed to get release info %s, %v", releasesUri, err)
		return
	}
	h.config.Set(func(c *config.Config) {
		c.ReleasesEtag = etag
	})

	var releases []Release
	if len(body) > 0 {
		err = json.Unmarshal(body, &releases)
		if err != nil {
			log.Warnf("Failed to parse release info, %v", err)
			return
		}
		preRelease, ok := h.getPreRelease(releases)
		if ok {
			h.config.Set(func(c *config.Config) {
				c.PreRelease = h.toConfigRelease(preRelease)
			})
		}
		stableRelease, ok := h.getStableRelease(releases)
		if ok {
			h.config.Set(func(c *config.Config) {
				c.StableRelease = h.toConfigRelease(stableRelease)
			})
		}
	}

	log.Infof("Response: %s, %s", etag, string(body))
}

func (h *autoUpdate) httpGet(url, requestEtag string) (bytes []byte, etag string, err error) {
	log.Infof("HTTP GET %s ...", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(log.Error(err))
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

	etagValue := resp.Header["Etag"] // W/"48e707ef72c8aa13423a1116ca40cbdd"
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
	var assets []config.Asset
	for _, a := range release.Assets {
		assets = append(assets, config.Asset{
			Name: a.Name,
			Url:  a.Browser_download_url,
		})
	}
	return config.Release{
		Version: release.Tag_name,
		Assets:  assets,
	}
}
