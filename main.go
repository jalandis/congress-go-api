package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/jalandis/congress-go-api/api"
	"github.com/jalandis/congress-go-api/cache"
)

// Application configuration.
type Configuration struct {
	Port         int
	EmberPath    string
	ApiBase      string
	ApiKey       string
	CacheTimeOut string
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

	upcomingBillsUrl := "/api/v1/legislation"
	router.HandleFunc(upcomingBillsUrl, client.HandleUpcomingBills)

	billUrl := "/api/v1/legislation/{billId}-{congressId:[0-9]+}"
	router.HandleFunc(billUrl, client.HandleBill)

	cosponsorsUrl := "/api/v1/legislation/{billId}-{congressId:[0-9]+}/representatives"
	router.HandleFunc(cosponsorsUrl, client.HandleBillCosponsors)

	statementsUrl := "/api/v1/legislation/{billId}-{congressId:[0-9]+}/statements"
	router.HandleFunc(statementsUrl, client.HandleBillStatements)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir(config.EmberPath)))

	headers := handlers.AllowedHeaders([]string{"*"})
	methods := handlers.AllowedMethods([]string{"GET"})
	origins := handlers.AllowedOrigins([]string{"*"})
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(
		fmt.Sprintf(":%d", config.Port),
		handlers.CORS(headers, methods, origins)(router),
	))
}
