package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ProxiesStore interface {
	GetProxyDetails(string) *url.URL
}

type ConfigServer struct {
	store ProxiesStore
}

func NewConfigServer(store ProxiesStore) *ConfigServer {
	return &ConfigServer{store}
}

func (c *ConfigServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := strings.TrimPrefix(r.URL.Path, "/proxies/")

	details := c.store.GetProxyDetails(proxy).String()
	fmt.Fprint(w, details)

	// fmt.Fprint(w, "http://proxy0:1000\nhttp://proxy1:1001")
}
