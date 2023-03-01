package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/lo"
)

type ProxyUrl = *url.URL

type ProxiesStore interface {
	GetProxyDetails(string) (ProxyUrl, bool)
	SetProxy(string, ProxyUrl) error
	GetAll() []ProxyUrl
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
	case http.MethodPut:
		c.processUpdateProxy(w, r)
	case http.MethodGet:
		c.processGetProxy(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func parseBodyAsUrl(w http.ResponseWriter, r *http.Request) (ProxyUrl, bool) {
	buf, _ := io.ReadAll(r.Body)
	proxy := string(buf)
	proxyUrl, err := url.ParseRequestURI(proxy)
	if err != nil {
		sendBadRequest(w, "Malformed Proxy URL")
		return nil, false
	}

	return proxyUrl, true
}

func (c *ConfigServer) processNewProxy(w http.ResponseWriter, r *http.Request) {
	proxyUrl, ok := parseBodyAsUrl(w, r)
	if !ok {
		return
	}

	host := proxyUrl.Hostname()
	_, listed := c.store.GetProxyDetails(host)
	if listed {
		sendBadRequest(w, "Proxy Already Listed")
		return
	}

	c.store.SetProxy(host, proxyUrl)
	w.WriteHeader(http.StatusAccepted)
}

func (c *ConfigServer) processUpdateProxy(w http.ResponseWriter, r *http.Request) {
	host := strings.TrimPrefix(r.URL.Path, "/proxies/")

	proxyUrl, ok := parseBodyAsUrl(w, r)
	if !ok {
		return
	}

	hostFromBody := proxyUrl.Hostname()

	if hostFromBody != host {
		sendBadRequest(w, "Request path doesn't match URL in body")
		return
	}

	c.store.SetProxy(host, proxyUrl)
	w.WriteHeader(http.StatusAccepted)
}

func sendBadRequest(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, msg)
}

func (c *ConfigServer) processGetProxy(w http.ResponseWriter, r *http.Request) {
	shouldGetAll := strings.HasSuffix(strings.TrimSuffix(r.URL.Path, "/"), "/proxies")
	if shouldGetAll {
		allProxyUrls := c.store.GetAll()
		all := strings.Join(lo.Map(allProxyUrls, func(u ProxyUrl, _ int) string { return u.String() }), "\n")
		fmt.Fprint(w, all)
		return
	}

	host := strings.TrimPrefix(r.URL.Path, "/proxies/")
	proxyUrl, ok := c.store.GetProxyDetails(host)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprint(w, proxyUrl.String())
}
