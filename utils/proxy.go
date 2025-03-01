package utils

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

//adding proxy to jsonrpc client

func NewHTTPClientWithProxy(proxyURL string) (*http.Client, error) {
	if proxyURL == "" {
		return &http.Client{}, nil // Return a default client if no proxy is specified
	}

	proxy, err := url.Parse(proxyURL)
	if err != nil {
		log.Println("error parsing proxy url", err.Error())
		return nil, err // Return an error if the proxy URL is invalid
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second}, nil
}
