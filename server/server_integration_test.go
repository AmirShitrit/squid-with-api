package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type InMemoryProxiesStore struct {
	proxies map[string]ProxyUrl
	lock    sync.Mutex
}

func (i *InMemoryProxiesStore) GetProxyDetails(host string) (ProxyUrl, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()
	proxy, ok := i.proxies[host]
	return proxy, ok
}

func (i *InMemoryProxiesStore) GetAll() []ProxyUrl {
	i.lock.Lock()
	defer i.lock.Unlock()
	return maps.Values(i.proxies)
}

func (i *InMemoryProxiesStore) SetProxy(host string, proxy ProxyUrl) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.proxies[host] = proxy
	return nil
}

func TestRegisteringProxiesAndRetrievingThem(t *testing.T) {
	store := &InMemoryProxiesStore{make(map[string]ProxyUrl), sync.Mutex{}}
	server := NewConfigServer(store)

	wantedProxies := lo.Map(
		lo.RangeFrom(0, 50),
		func(item int, _ int) string {
			padded := fmt.Sprintf("1%03d", item)
			return fmt.Sprintf("http://user%[1]s:pwd%[1]s@proxy%[1]s:1%[1]s", padded)
		})

	w := sync.WaitGroup{}
	for _, proxy := range wantedProxies {
		w.Add(1)
		proxy := proxy
		go func() {
			server.ServeHTTP(httptest.NewRecorder(), newPostProxyRequest(proxy))
			w.Done()
		}()
	}
	w.Wait()

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetAllRequest())

	assertResponseStatus(t, response.Code, http.StatusOK)
	gotProxies := strings.Split(response.Body.String(), "\n")
	slices.Sort(gotProxies)
	assertResponseBody(t, gotProxies, wantedProxies)
}
