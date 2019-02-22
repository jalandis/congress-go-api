# Go Wrapper for ProPublica Congress API [![CircleCI](https://circleci.com/gh/jalandis/congress-go-api/tree/master.svg?style=svg)](https://circleci.com/gh/jalandis/congress-go-api/tree/master)

Backend API that wraps the [ProPublic Congress API](https://www.propublica.org/datastore/api/propublica-congress-api).

* Handles API Key securely
* Caches API requests for 24 hours
* Converts ProPublica API responses to jsonapi format
* Serves Ember site

## Development Environment Setup

Dependencies:
* Go 1.10.4
* dep v0.5.0

A [ProPublica API key](https://www.propublica.org/datastore/api/propublica-congress-api) will be required.  After registering with the Propublica Data Store, an email with the key will be sent.

The application has its settings passed in through a [JSON configuration file](https://github.com/jalandis/congress-go-api/blob/master/config_template.json).  Below is an example of the required format.  Please alter paths, ports and API key to fit your environment.  See the [Ember Frontend](#ember-frontend) section for more information on the **EmberPath** variable.

    {
        "Port": 8080,
        "EmberPath": "***/congress-ember/dist",
        "ApiBase": "https://api.propublica.org/congress/v1",
        "ApiKey": "***",
        "CacheTimeOut": "24h"
    }

Install Dependencies:

    go get ./..

Alternative Installation of Dependencies (still in testing):

    dep ensure

Run Server:

    go run github.com/jalandis/congress/main.go --config_path ./config.json

### Ember FrontEnd

This API project serves a [Congress Ember application](https://github.com/jalandis/congress-ember) and will require the Ember app be built locally for testing, development, or demo.

Please follow the **installation** and **build** sections of the [Ember README](https://github.com/jalandis/congress-ember/blob/master/README.md) and update the **EmberPath** value in this projects json configuration to point to the Ember dist folder.

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
