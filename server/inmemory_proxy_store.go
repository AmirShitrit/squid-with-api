package server

import (
	"sync"

	"golang.org/x/exp/maps"
)

type InMemoryProxiesStore struct {
	proxies map[string]ProxyUrl
	lock    sync.Mutex
}

func NewInMemoryProxiesStore() *InMemoryProxiesStore {
	return &InMemoryProxiesStore{make(map[string]ProxyUrl), sync.Mutex{}}
}

func (i *InMemoryProxiesStore) GetProxyDetails(host string) (ProxyUrl, bool, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	proxy, ok := i.proxies[host]
	return proxy, ok, nil
}

func (i *InMemoryProxiesStore) GetAll() ([]ProxyUrl, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	return maps.Values(i.proxies), nil
}

func (i *InMemoryProxiesStore) SetProxy(host string, proxy ProxyUrl) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.proxies[host] = proxy
	return nil
}
