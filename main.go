package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/AmirShitrit/squid-with-api/server"
)

type InMemoryProxiesStore struct{}

func (i *InMemoryProxiesStore) GetProxyDetails(host string) (*url.URL, bool) {
	return &url.URL{Scheme: "http", Host: "proxy9:1009"}, true
}

func (i *InMemoryProxiesStore) GetAll() []*url.URL {
	return []*url.URL{{Scheme: "http", Host: "proxy9:1009"}}
}

func (i *InMemoryProxiesStore) SetProxy(string, *url.URL) error {
	return nil
}

const port = 5000

func main() {
	server := server.NewConfigServer(&InMemoryProxiesStore{})
	log.Printf("Listening on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 5000), server))
}
