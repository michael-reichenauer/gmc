package tools

import (
	"bytes"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"

	"io/ioutil"
	"net/http"
)

func httpGet(url string) (bytes []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(log.Fatal(err))
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("user-agent", "gmc")

	client := &http.Client{Transport: &http.Transport{Proxy: utils.GetHTTPProxy()}}
	resp, err := client.Do(req)
	if err != nil {
		ErrorEcho("No contact with %s, %s", url, err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		ErrorEcho("Invalid status for %s: %d", url, resp.StatusCode)
		return
	}

	bytes, err = ioutil.ReadAll(resp.Body)
	return
}

func httpPost(url string, body []byte) (data []byte, err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		panic(log.Fatal(err))
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("user-agent", "gmc")

	client := &http.Client{Transport: &http.Transport{Proxy: nil}}
	resp, err := client.Do(req)
	if err != nil {
		ErrorEcho("No contact with %s, %s", url, err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		ErrorEcho("Invalid status for %s: %d", url, resp.StatusCode)
		return
	}

	data, err = ioutil.ReadAll(resp.Body)
	return
}
