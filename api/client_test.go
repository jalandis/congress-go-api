package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/jalandis/congress-go-api/api"
	"github.com/jalandis/congress-go-api/cache"
)

func setupTest(t *testing.T) (*httptest.Server, api.ApiClient) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		// All requests are required to have an API key.
		key := req.Header.Get("X-API-Key")
		if key != "api-key" {
			t.Errorf("API request missing required key")
		}

		path := fmt.Sprintf("./test_data%s", req.URL.Path)
		http.ServeFile(rw, req, path)
	}))

	base, err := url.Parse(server.URL)
	if err != nil {
		t.Errorf("Failed setting up server in TestApi: %s", err)
	}

	ttl, _ := time.ParseDuration("24h")
	return server, api.ApiClient{
		Key:  "api-key",
		Base: base,
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
		Cache: cache.New(),
		Ttl:   ttl,
	}
}

func TestApi(t *testing.T) {

	t.Run("Test API Wrappers", func(t *testing.T) {
		t.Run("Test API Upcoming Bills", func(t *testing.T) {
			t.Parallel()

			server, apiClient := setupTest(t)
			defer server.Close()

			result, err := apiClient.GetUpcomingBills("house")
			if err != nil {
				t.Errorf("Error requesting upcoming bills: %s", err)
			}

			if len(result) <= 0 {
				t.Errorf("API returned a bad response")
			}
		})

		t.Run("Test API Cosponsors", func(t *testing.T) {
			t.Parallel()

			server, apiClient := setupTest(t)
			defer server.Close()

			result, err := apiClient.GetBillCosponsers(115, "hr4249")
			if err != nil {
				t.Errorf("Error requesting bill cosponsors: %s", err)
			}

			if len(result) <= 0 {
				t.Errorf("API returned a bad response")
			}
		})

		t.Run("Test API caching", func(t *testing.T) {
			t.Parallel()

			server, apiClient := setupTest(t)

			result, err := apiClient.GetBillCosponsers(115, "hr4249")
			if err != nil {
				t.Errorf("Error requesting bill cosponsors: %s", err)
			}

			if len(result) <= 0 {
				t.Errorf("API returned a bad response")
			}

			// Closing server to confirm no more requests sent.
			server.Close()

			result, err = apiClient.GetBillCosponsers(115, "hr4249")
			if err != nil {
				t.Errorf("Error requesting bill cosponsors: %s", err)
			}

			if len(result) <= 0 {
				t.Errorf("API returned a bad response")
			}
		})
	})
}
