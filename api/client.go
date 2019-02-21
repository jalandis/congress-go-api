package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jalandis/congress-go-api/cache"
	log "github.com/sirupsen/logrus"
)

var (
	expectedStatus = "OK"
	noDataMessage  = "No data returned from ProPublic API"
	badStatus      = "Bad status reported from ProPublic API: %s"
)

// ProPublica API client.
type ApiClient struct {
	Key    string
	Base   *url.URL
	Client *http.Client
	Cache  cache.Cache
	Ttl    time.Duration
}

// ProPublica response to upcoming bills API get request.
type UpcomingBillsApiResponse struct {
	Status  string `json:"status"`
	Results []struct {
		Bills []*UpcomingBill `json:"bills"`
	} `json:"results"`
}

// Bill object with jsonapi formatting.
type UpcomingBill struct {
	Id       string `json:"bill_id" jsonapi:"primary,legislation"`
	Name     string `json:"description" jsonapi:"attr,name"`
	Number   string `json:"bill_number" jsonapi:"attr,number"`
	Slug     string `json:"bill_slug" jsonapi:"attr,slug"`
	Chamber  string `json:"chamber" jsonapi:"attr,chamber"`
	Congress string `json:"congress" jsonapi:"attr,congress"`
	Url      string `json:"bill_url" jsonapi:"attr,url"`
}

// API request for upcoming bills to {house|senate}.
func (api ApiClient) GetUpcomingBills(chamberId string) ([]*UpcomingBill, error) {
	endpoint := fmt.Sprintf("%s/bills/upcoming/%s.json", api.Base, chamberId)
	if value, ok := api.Cache.Get(endpoint); ok {
		cachedResult, _ := value.([]*UpcomingBill)
		return cachedResult, nil
	}

	response := new(UpcomingBillsApiResponse)
	if err := api.request(endpoint, response); err != nil {
		return nil, err
	}

	if response.Status != expectedStatus {
		return nil, errors.New(fmt.Sprintf(badStatus, response.Status))
	}

	if len(response.Results) == 0 {
		return nil, errors.New(noDataMessage)
	}

	api.Cache.Set(endpoint, response.Results[0].Bills, api.Ttl)
	return response.Results[0].Bills, nil
}

// Sparse ProPublica response to bill sponsors API get request.
type BillCosponsorsApiResponse struct {
	Status  string `json:"status"`
	Results []struct {
		Cosponsors []*Representative `json:"cosponsors"`
	} `json:"results"`
}

// Sparse Representative object with jsonapi formatting.
type Representative struct {
	Id    string `json:"cosponsor_id" jsonapi:"primary,representative"`
	Name  string `json:"name" jsonapi:"attr,name"`
	Party string `json:"cosponsor_party" jsonapi:"attr,party-id"`
	State string `json:"cosponsor_state" jsonapi:"attr,state"`
}

// API request for bills sponsors.
func (api ApiClient) GetBillCosponsers(congressId int, billId string) ([]*Representative, error) {
	endpoint := fmt.Sprintf("%s/%d/bills/%s/cosponsors.json", api.Base, congressId, billId)
	if value, ok := api.Cache.Get(endpoint); ok {
		cachedResult, _ := value.([]*Representative)
		return cachedResult, nil
	}

	response := new(BillCosponsorsApiResponse)
	if err := api.request(endpoint, response); err != nil {
		return nil, err
	}

	if response.Status != expectedStatus {
		return nil, errors.New(fmt.Sprintf(badStatus, response.Status))
	}

	if len(response.Results) == 0 {
		return nil, errors.New(noDataMessage)
	}

	api.Cache.Set(endpoint, response.Results[0].Cosponsors, api.Ttl)
	return response.Results[0].Cosponsors, nil
}

// ProPublica response to bill statements API get request.
type StatementsApiResponse struct {
	Status  string       `json:"status"`
	Results []*Statement `json:"results"`
}

// Sparse Statements object with jsonapi formatting.
type Statement struct {
	Id      string `json:"url" jsonapi:"primary,statement"`
	Title   string `json:"title" jsonapi:"attr,title"`
	Type    string `json:"type" jsonapi:"attr,type"`
	Speaker string `json:"name" jsonapi:"attr,speaker"`
}

// API request for bills sponsors.
func (api ApiClient) GetBillStatements(congressId int, billSlug string) ([]*Statement, error) {
	endpoint := fmt.Sprintf("%s/%d/bills/%s/statements.json", api.Base, congressId, billSlug)
	if value, ok := api.Cache.Get(endpoint); ok {
		cachedResult, _ := value.([]*Statement)
		return cachedResult, nil
	}

	response := new(StatementsApiResponse)
	if err := api.request(endpoint, response); err != nil {
		return nil, err
	}

	if response.Status != expectedStatus {
		return nil, errors.New(fmt.Sprintf(badStatus, response.Status))
	}

	api.Cache.Set(endpoint, response.Results, api.Ttl)
	return response.Results, nil
}

// ProPublica response to specific bill API get request.
type BillApiResponse struct {
	Status  string  `json:"status"`
	Results []*Bill `json:"results"`
}

// Bill object with jsonapi formatting.
type Bill struct {
	Id       string `json:"bill_id" jsonapi:"primary,legislation"`
	Name     string `json:"short_title" jsonapi:"attr,name"`
	Number   string `json:"number" jsonapi:"attr,number"`
	Slug     string `json:"bill_slug" jsonapi:"attr,slug"`
	Congress string `json:"congress" jsonapi:"attr,congress"`
	Url      string `json:"gpo_pdf_uri" jsonapi:"attr,url"`
}

// API request for a specific bill.
func (api ApiClient) GetBill(congressId int, billSlug string) (*Bill, error) {
	endpoint := fmt.Sprintf("%s/%d/bills/%s.json", api.Base, congressId, billSlug)
	if value, ok := api.Cache.Get(endpoint); ok {
		cachedResult, _ := value.(*Bill)
		return cachedResult, nil
	}

	response := new(BillApiResponse)
	if err := api.request(endpoint, response); err != nil {
		return nil, err
	}

	if response.Status != expectedStatus {
		return nil, errors.New(fmt.Sprintf(badStatus, response.Status))
	}

	if len(response.Results) == 0 {
		return nil, errors.New(noDataMessage)
	}

	api.Cache.Set(endpoint, response.Results[0], api.Ttl)
	return response.Results[0], nil
}

// Wrapper for ProPublica API request injecting API key.
func (api ApiClient) request(endpoint string, result interface{}) error {
	log.WithFields(log.Fields{"endpoint": endpoint}).Info("Calling Propublic API")

	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("X-API-Key", api.Key)
	if resp, err := api.Client.Do(req); err != nil {
		return err
	} else {
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.WithFields(log.Fields{"err": err}).Info("Failed decoding response")
			return err
		}
	}

	log.WithFields(log.Fields{"result": result}).Debug("Results from Propublic API")
	return nil
}
