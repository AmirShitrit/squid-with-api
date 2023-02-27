package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ProxiesStore interface {
	GetProxyDetails(string) (*url.URL, bool)
}

type ConfigServer struct {
	store ProxiesStore
}

func NewConfigServer(store ProxiesStore) *ConfigServer {
	return &ConfigServer{store}
}

func (c *ConfigServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := strings.TrimPrefix(r.URL.Path, "/proxies/")

	details, ok := c.store.GetProxyDetails(proxy)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprint(w, details.String())
}
