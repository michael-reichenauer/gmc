package utils

import (
	"github.com/rapid7/go-get-proxied/proxy"
	"net"
	"net/http"
	"net/url"
	"time"
)

func SetDefaultHTTPProxy() {
	provider := proxy.NewProvider("")
	proxyURL := provider.GetHTTPProxy("https://api.github.com")
	if proxyURL == nil {
		// No proxy needed
		return
	}

	defaultTransport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL.URL()),
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 15 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	http.DefaultClient.Transport = defaultTransport
}

func GetHTTPProxy() func(*http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		provider := proxy.NewProvider("")
		proxyURL := provider.GetHTTPProxy(req.URL.String())
		if proxyURL == nil {
			// No proxy needed
			return nil, nil
		}
		return proxyURL.URL(), nil
	}
}

func GetHTTPProxyURL() string {
	provider := proxy.NewProvider("")
	proxyURL := provider.GetHTTPProxy("https://api.github.com")
	if proxyURL == nil {
		// No proxy needed
		return ""
	}
	return proxyURL.URL().String()
}
