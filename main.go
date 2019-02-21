package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/jsonapi"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"jalandis.com/congress/api"
	"jalandis.com/congress/cache"
)

// Application configuration.
type Configuration struct {
	Port         int
	Public       string
	ApiBase      string
	ApiKey       string
	CacheTimeOut string
}

// Converts ProPublica upcoming bills API endoint to jsonapi format for use
// by Ember application.
//
// Combines upcoming bills from both the house and senate.
func HandleUpcomingBills(client api.ApiClient, rw http.ResponseWriter, req *http.Request) {
	var wg sync.WaitGroup
	wg.Add(2)

	var houseResults []*api.UpcomingBill
	var houseError error
	var senateResults []*api.UpcomingBill
	var senateError error

	go func() {
		defer wg.Done()
		houseResults, houseError = client.GetUpcomingBills("house")
	}()

	go func() {
		defer wg.Done()
		senateResults, senateError = client.GetUpcomingBills("senate")
	}()
	wg.Wait()

	if houseError != nil {
		http.Error(rw, houseError.Error(), http.StatusInternalServerError)
	}

	if senateError != nil {
		http.Error(rw, senateError.Error(), http.StatusInternalServerError)
	}

	allBills := append(houseResults, senateResults...)
	rw.Header().Set("Content-Type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(rw, allBills); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// Converts ProPublica bill sponsors API endoint to jsonapi format for use
// by Ember application.
func HandleBillCosponsors(client api.ApiClient, rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	congressId, _ := strconv.Atoi(vars["congressId"])
	result, apiError := client.GetBillCosponsers(congressId, vars["billId"])
	if apiError != nil {
		http.Error(rw, apiError.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(rw, result); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// Converts ProPublica bill statements API endoint to jsonapi format for use
// by Ember application.
func HandleBillStatements(client api.ApiClient, rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	congressId, _ := strconv.Atoi(vars["congressId"])
	result, apiError := client.GetBillStatements(congressId, vars["billSlug"])
	if apiError != nil {
		http.Error(rw, apiError.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(rw, result); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// Converts ProPublica bill API endoint to jsonapi format for use by Ember application.
func HandleBill(client api.ApiClient, rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	parts := strings.Split(vars["billSlug"], "-")
	congressId, _ := strconv.Atoi(parts[1])
	result, apiError := client.GetBill(congressId, parts[0])
	if apiError != nil {
		http.Error(rw, apiError.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(rw, result); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	configPath := flag.String("config_path", "", "Path to configuration file.")
	flag.Parse()

	file, err := os.Open(*configPath)
	if err != nil {
		panic(err)
	}

	config := Configuration{}
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		panic(err)
	}

	url, _ := url.Parse(config.ApiBase)
	ttl, _ := time.ParseDuration(config.CacheTimeOut)
	client := api.ApiClient{
		Key:  config.ApiKey,
		Base: url,
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
		Cache: cache.New(),
		Ttl:   ttl,
	}

	router := mux.NewRouter()

	upcomingBillsUrl := "/congress/v1/legislation"
	router.HandleFunc(upcomingBillsUrl, func(rw http.ResponseWriter, req *http.Request) {
		HandleUpcomingBills(client, rw, req)
	})

	billUrl := "/congress/v1/legislation/{billSlug}"
	router.HandleFunc(billUrl, func(rw http.ResponseWriter, req *http.Request) {
		HandleBill(client, rw, req)
	})

	cosponsorsUrl := "/congress/v1/legislation/{billId}-{congressId:[0-9]+}/representatives"
	router.HandleFunc(cosponsorsUrl, func(rw http.ResponseWriter, req *http.Request) {
		HandleBillCosponsors(client, rw, req)
	})

	statementsUrl := "/congress/v1/congress/{congressId:[0-9]+}/legislation/{billSlug}/statements"
	router.HandleFunc(statementsUrl, func(rw http.ResponseWriter, req *http.Request) {
		HandleBillStatements(client, rw, req)
	})

	router.PathPrefix("/").Handler(http.FileServer(http.Dir(config.Public)))

	headers := handlers.AllowedHeaders([]string{"*"})
	methods := handlers.AllowedMethods([]string{"GET"})
	origins := handlers.AllowedOrigins([]string{"*"})
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(
		fmt.Sprintf(":%d", config.Port),
		handlers.CORS(headers, methods, origins)(router),
	))
}
