package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"sync"
	"testing"

	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

func randomDbFilePath() string {
	return path.Join(os.TempDir(), fmt.Sprintf("proxies_%s.boltdb", lo.RandomString(8, lo.LettersCharset)))
}

func TestRegisteringProxiesAndRetrievingThem(t *testing.T) {
	path := randomDbFilePath()
	store, err := NewBoltStorage(path)
	if err != nil {
		t.Fatal(err)
	}

	defer store.Close()
	defer os.Remove(path)

	server := NewConfigServer(store)

	t.Run("register many proxies and retrieve them all", func(t *testing.T) {
		wantedProxies := lo.Map(
			lo.RangeFrom(0, 20),
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
	})
}
