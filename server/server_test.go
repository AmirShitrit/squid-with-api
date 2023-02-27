package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

type StubProxiesStore struct {
	proxies map[string]*url.URL
}

func (s *StubProxiesStore) GetProxyDetails(host string) *url.URL {
	details := s.proxies[host]
	return details
}

func TestGetProxies(t *testing.T) {
	t.Run("playground", func(t *testing.T) {
		t.Skip()
	})

	stubProxiesStore := StubProxiesStore{
		map[string]*url.URL{
			"proxy0": {Scheme: "http", Host: "proxy0:1000"},
			"proxy1": {Scheme: "http", Host: "proxy1:1001"},
		},
	}
	server := &ConfigServer{store: &stubProxiesStore}

	t.Run("get proxy0's details", func(t *testing.T) {
		request := newGetProxyDetailsRequest("proxy0")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Body.String()
		want := "http://proxy0:1000"
		assertResponseBody(t, got, want)
	})

	t.Run("get proxy1's details", func(t *testing.T) {
		request := newGetProxyDetailsRequest("proxy1")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Body.String()
		want := "http://proxy1:1001"
		assertResponseBody(t, got, want)
	})

	// t.Run("returns all proxies", func(t *testing.T) {
	// 	request, _ := http.NewRequest(http.MethodGet, "/proxies", nil)
	// 	response := httptest.NewRecorder()

	// 	ConfigServer(response, request)

	// 	body := response.Body.String()
	// 	got := strings.Split(body, "\n")
	// 	slices.Sort(got)
	// 	want := []string{
	// 		"http://proxy0:1000",
	// 		"http://proxy1:1001",
	// 		// "http://user2:password2@proxy2:1002",
	// 		// "http://user3:password3@proxy3:1003",
	// 	}
	// 	slices.Sort(want)
	// 	assertResponseBody(t, got, want)
	// })
}

func newGetProxyDetailsRequest(proxy string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/proxies/%s", proxy), nil)
	return req
}

func assertResponseBody(t testing.TB, got, want any) {
	t.Helper()

	gotWhatWant := reflect.DeepEqual(got, want)

	if !gotWhatWant {
		t.Errorf("got %q, want %q", got, want)
	}
}
