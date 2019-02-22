package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jalandis/congress-go-api/api"
	"github.com/jalandis/congress-go-api/cache"

	"github.com/gorilla/mux"
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

	t.Run("Test ProPublica API Wrappers", func(t *testing.T) {
		t.Run("Test Get Upcoming Bills", func(t *testing.T) {
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

		t.Run("Test Get Cosponsors", func(t *testing.T) {
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

		t.Run("Test Get Statements", func(t *testing.T) {
			t.Parallel()

			server, apiClient := setupTest(t)
			defer server.Close()

			result, err := apiClient.GetBillStatements(115, "s19")
			if err != nil {
				t.Errorf("Error requesting bill statements: %s", err)
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

		t.Run("Test Handling Bill Cosponsor Request", func(t *testing.T) {
			t.Parallel()
			_, apiClient := setupTest(t)

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, map[string]string{
				"congressId": "115",
				"billId":     "hr4249",
			})

			apiClient.HandleBillCosponsors(rw, req)

			if rw.Code != http.StatusOK {
				t.Errorf("Expected status OK. Got: %d", rw.Code)
			}

			results := rw.Body.String()
			if !strings.HasPrefix(results, "{\"data\":[{") {
				t.Errorf("Expected data results: %s", results)
			}
		})

		t.Run("Test Handling Bill Statement Request", func(t *testing.T) {
			t.Parallel()
			_, apiClient := setupTest(t)

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, map[string]string{
				"congressId": "115",
				"billId":     "s19",
			})

			apiClient.HandleBillStatements(rw, req)

			if rw.Code != http.StatusOK {
				t.Errorf("Expected status OK. Got: %d", rw.Code)
			}

			results := rw.Body.String()
			if !strings.HasPrefix(results, "{\"data\":[{") {
				t.Errorf("Expected data results: %s", results)
			}
		})

		t.Run("Test Handling Upcoming Bills Request", func(t *testing.T) {
			t.Parallel()
			_, apiClient := setupTest(t)

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			apiClient.HandleUpcomingBills(rw, req)

			if rw.Code != http.StatusOK {
				t.Errorf("Expected status OK. Got: %d", rw.Code)
			}

			results := rw.Body.String()
			if !strings.HasPrefix(results, "{\"data\":[{") {
				t.Errorf("Expected data results: %s", results)
			}
		})
	})
}
