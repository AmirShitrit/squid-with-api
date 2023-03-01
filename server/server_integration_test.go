package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type InMemoryProxiesStore struct {
	proxies map[string]ProxyUrl
}

func (i *InMemoryProxiesStore) GetProxyDetails(host string) (ProxyUrl, bool) {
	proxy, ok := i.proxies[host]
	return proxy, ok
}

func (i *InMemoryProxiesStore) GetAll() []ProxyUrl {
	return maps.Values(i.proxies)
}

func (i *InMemoryProxiesStore) SetProxy(host string, proxy ProxyUrl) error {
	i.proxies[host] = proxy
	return nil
}

func TestRegisteringProxiesAndRetrievingThem(t *testing.T) {
	store := &InMemoryProxiesStore{make(map[string]ProxyUrl)}
	server := NewConfigServer(store)

	wantedProxies := []string{
		"http://user1:pwd1@proxy1:1001",
		"http://user2:pwd2@proxy2:1002",
		"http://user3:pwd3@proxy3:1003",
	}
	for _, proxy := range wantedProxies {
		server.ServeHTTP(httptest.NewRecorder(), newPostProxyRequest(proxy))
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetAllRequest())

	assertResponseStatus(t, response.Code, http.StatusOK)
	gotProxies := strings.Split(response.Body.String(), "\n")
	slices.Sort(gotProxies)
	assertResponseBody(t, gotProxies, wantedProxies)
}
