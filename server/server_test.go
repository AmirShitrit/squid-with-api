package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type StubProxiesStore struct {
	proxies map[string]ProxyUrl
}

func (s *StubProxiesStore) GetProxyDetails(host string) (ProxyUrl, bool, error) {
	proxyUrl, ok := s.proxies[host]
	return proxyUrl, ok, nil
}

func (s *StubProxiesStore) GetAll() ([]ProxyUrl, error) {
	return maps.Values(s.proxies), nil
}

func (s *StubProxiesStore) SetProxy(host string, proxyUrl ProxyUrl) error {
	s.proxies[host] = proxyUrl
	return nil
}

func TestGetProxies(t *testing.T) {
	stubProxiesStore := StubProxiesStore{
		map[string]ProxyUrl{
			"proxy0": {Scheme: "http", Host: "proxy0:1000"},
			"proxy1": {Scheme: "http", Host: "proxy1:1001"},
		},
	}
	server := &ConfigServer{store: &stubProxiesStore}

	t.Run("get proxy0's details", func(t *testing.T) {
		request := newGetProxyDetailsRequest("proxy0")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseStatus(t, response.Code, http.StatusOK)

		got := response.Body.String()
		want := "http://proxy0:1000"
		assertResponseBody(t, got, want)
	})

	t.Run("get proxy1's details", func(t *testing.T) {
		request := newGetProxyDetailsRequest("proxy1")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseStatus(t, response.Code, http.StatusOK)

		got := response.Body.String()
		want := "http://proxy1:1001"
		assertResponseBody(t, got, want)
	})

	t.Run("returns 404 on missing proxies", func(t *testing.T) {
		request := newGetProxyDetailsRequest("non-existing-proxy")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Code
		want := http.StatusNotFound

		assertResponseBody(t, got, want)
	})

	t.Run("returns all proxies", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/proxies", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		body := response.Body.String()
		got := strings.Split(body, "\n")
		slices.Sort(got)
		want := []string{
			"http://proxy0:1000",
			"http://proxy1:1001",
			// "http://user2:password2@proxy2:1002",
			// "http://user3:password3@proxy3:1003",
		}
		slices.Sort(want)
		assertResponseBody(t, got, want)
	})
}

func TestUpdateProxies(t *testing.T) {
	store := StubProxiesStore{
		map[string]ProxyUrl{
			"proxy0": {Scheme: "http", Host: "proxy0:1000"},
		},
	}
	server := &ConfigServer{&store}

	t.Run("it returns accepted on PUT", func(t *testing.T) {
		newUrl := "http://proxy0:10001"
		request := newPutProxyRequest(newUrl)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseStatus(t, response.Code, http.StatusAccepted)
		got := store.proxies["proxy0"].String()
		want := newUrl
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})
}

func TestAddProxies(t *testing.T) {
	store := StubProxiesStore{
		map[string]ProxyUrl{},
	}
	server := &ConfigServer{&store}

	t.Run("it returns accepted on POST", func(t *testing.T) {
		request := newPostProxyRequest("http://proxy3:1003")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseStatus(t, response.Code, http.StatusAccepted)

		if len(store.proxies) != 1 {
			t.Fatalf("Expected 1 proxy, but got %d", len(store.proxies))
		}

		if _, ok := store.proxies["proxy3"]; !ok {
			t.Errorf("proxy3 was not saved")
		}
	})

	t.Run("it returns 400 Bad Request on malformed URL", func(t *testing.T) {
		request := newPostProxyRequest("not a URL:")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertResponseStatus(t, response.Code, http.StatusBadRequest)

		got := response.Body.String()
		want := "Malformed Proxy URL"
		assertResponseBody(t, got, want)
	})

	t.Run("it returns 400 Bad Request on proxy already exists", func(t *testing.T) {
		server.ServeHTTP(httptest.NewRecorder(), newPostProxyRequest("http://proxy1:1001"))

		response := httptest.NewRecorder()
		server.ServeHTTP(response, newPostProxyRequest("http://proxy1:1001"))

		assertResponseStatus(t, response.Code, http.StatusBadRequest)
		got := response.Body.String()
		want := "Proxy Already Listed"
		assertResponseBody(t, got, want)
	})
}

func newGetProxyDetailsRequest(proxy string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/proxies/%s", proxy), nil)
	return req
}

func newGetAllRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/proxies", nil)
	return req
}

func newPutProxyRequest(proxy string) *http.Request {
	withoutSchema := strings.TrimPrefix(proxy, "http://")
	hostName := withoutSchema[0:strings.Index(withoutSchema, ":")]
	body := strings.NewReader(proxy)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/proxies/%s", hostName), body)
	return req
}

func newPostProxyRequest(proxy string) *http.Request {
	body := strings.NewReader(proxy)
	req, _ := http.NewRequest(http.MethodPost, "/proxies", body)
	return req
}

func assertResponseStatus(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("got wrong status code: expected %d, but got %d", want, got)
	}
}

func assertResponseBody(t testing.TB, got, want any) {
	t.Helper()

	gotWhatWant := reflect.DeepEqual(got, want)

	if !gotWhatWant {
		t.Errorf("got %q, want %q", got, want)
	}
}
