package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/lo"
)

type ProxiesStore interface {
	GetProxyDetails(string) (*url.URL, bool)
	SetProxy(string, *url.URL) error
	GetAll() []*url.URL
}

type ConfigServer struct {
	store ProxiesStore
}

func NewConfigServer(store ProxiesStore) *ConfigServer {
	return &ConfigServer{store}
}

func (c *ConfigServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.processNewProxy(w, r)
	case http.MethodGet:
		c.processGetProxy(w, r)
	}
}

func (c *ConfigServer) processNewProxy(w http.ResponseWriter, r *http.Request) {
	buf, _ := io.ReadAll(r.Body)
	proxy := string(buf)
	proxyUrl, err := url.ParseRequestURI(proxy)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Malformed Proxy URL")
		return
	}

	host := proxyUrl.Hostname()
	_, listed := c.store.GetProxyDetails(host)
	if listed {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Proxy Already Listed")
		return
	}

	c.store.SetProxy(host, proxyUrl)
	w.WriteHeader(http.StatusAccepted)
}

func (c *ConfigServer) processGetProxy(w http.ResponseWriter, r *http.Request) {
	shouldGetAll := strings.HasSuffix(strings.TrimSuffix(r.URL.Path, "/"), "/proxies")
	if shouldGetAll {
		allProxyUrls := c.store.GetAll()
		all := strings.Join(lo.Map(allProxyUrls, func(u *url.URL, _ int) string { return u.String() }), "\n")
		fmt.Fprint(w, all)
		return
	}

	proxy := strings.TrimPrefix(r.URL.Path, "/proxies/")

	proxyUrl, ok := c.store.GetProxyDetails(proxy)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprint(w, proxyUrl.String())
}
