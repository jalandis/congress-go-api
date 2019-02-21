# Go Wrapper for ProPublica Congress API [![CircleCI](https://circleci.com/gh/jalandis/congress-go-api/tree/master.svg?style=svg)](https://circleci.com/gh/jalandis/congress-go-api/tree/master)

Backend API that wraps the [ProPublic Congress API](https://www.propublica.org/datastore/api/propublica-congress-api).

* Handles API Key securily
* Caches API requests for 24 hours
* Converts ProPublica API responses to jsonapi format
* Serves Ember site

## Development Environment Setup

Dependencies:
* Go 1.10.4
* dep v0.5.0

A [ProPublica API key](https://www.propublica.org/datastore/api/propublica-congress-api) will be required.  After registering with the Propublica Data Store, an email with the key will be sent.

The application has its configuration passed in through a JSON file.  Below is an example of the required format.  Please alter paths, ports and API key to fit your environment.

    {
        "Port": 8080,
        "Public": "/home/jalandis/workspace/nodejs/congress/dist",
        "ApiBase": "https://api.propublica.org/congress/v1",
        "ApiKey": "***",
        "CacheTimeOut": "24h"
    }

Install Dependencies:

    dep ensure

    go get ./..

Run Server:

    go run github.com/jalandis/congress/main.go --config_path ~/server.json

## Helpful Go Commands

Format all code and run unit tests:

    gofmt -s -w .
    go test -race ./...

Evaluate test coverage:

    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

View documentation:

    godoc -http=:6060

## Helpful CircleCi Commands

    circleci local execute --job build
