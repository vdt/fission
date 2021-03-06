/*
Copyright 2016 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/fission/fission"
	"github.com/fission/fission/fission/logdb"
)

type (
	API struct {
		FunctionStore
		HTTPTriggerStore
		TimeTriggerStore
		EnvironmentStore
		WatchStore
	}

	logDBConfig struct {
		httpURL  string
		username string
		password string
	}
)

func MakeAPI(rs *ResourceStore) *API {
	api := &API{
		FunctionStore:    FunctionStore{ResourceStore: *rs},
		HTTPTriggerStore: HTTPTriggerStore{ResourceStore: *rs},
		TimeTriggerStore: TimeTriggerStore{ResourceStore: *rs},
		EnvironmentStore: EnvironmentStore{ResourceStore: *rs},
		WatchStore:       WatchStore{ResourceStore: *rs},
	}
	return api
}

func (api *API) respondWithSuccess(w http.ResponseWriter, resp []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := w.Write(resp)
	if err != nil {
		// this will probably fail too, but try anyway
		api.respondWithError(w, err)
	}
}

func (api *API) respondWithError(w http.ResponseWriter, err error) {
	debug.PrintStack()
	code, msg := fission.GetHTTPError(err)
	log.Errorf("Error: %v: %v", code, msg)
	http.Error(w, msg, code)
}

func (api *API) getLogDBConfig(dbType string) logDBConfig {
	dbType = strings.ToUpper(dbType)
	// retrieve db auth config from the env
	url := os.Getenv(fmt.Sprintf("%s_URL", dbType))
	if url == "" {
		// set up default database url
		url = logdb.INFLUXDB_URL
	}
	username := os.Getenv(fmt.Sprintf("%s_USERNAME", dbType))
	password := os.Getenv(fmt.Sprintf("%s_PASSWORD", dbType))
	return logDBConfig{
		httpURL:  url,
		username: username,
		password: password,
	}
}

func (api *API) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\"message\": \"Fission API\", \"version\": \"0.1.0\"}\n")
}

func (api *API) Serve(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/", api.HomeHandler)

	r.HandleFunc("/v1/functions", api.FunctionApiList).Methods("GET")
	r.HandleFunc("/v1/functions", api.FunctionApiCreate).Methods("POST")
	r.HandleFunc("/v1/functions/{function}", api.FunctionApiGet).Methods("GET")
	r.HandleFunc("/v1/functions/{function}", api.FunctionApiUpdate).Methods("PUT")
	r.HandleFunc("/v1/functions/{function}", api.FunctionApiDelete).Methods("DELETE")

	r.HandleFunc("/v1/triggers/http", api.HTTPTriggerApiList).Methods("GET")
	r.HandleFunc("/v1/triggers/http", api.HTTPTriggerApiCreate).Methods("POST")
	r.HandleFunc("/v1/triggers/http/{httpTrigger}", api.HTTPTriggerApiGet).Methods("GET")
	r.HandleFunc("/v1/triggers/http/{httpTrigger}", api.HTTPTriggerApiUpdate).Methods("PUT")
	r.HandleFunc("/v1/triggers/http/{httpTrigger}", api.HTTPTriggerApiDelete).Methods("DELETE")

	r.HandleFunc("/v1/environments", api.EnvironmentApiList).Methods("GET")
	r.HandleFunc("/v1/environments", api.EnvironmentApiCreate).Methods("POST")
	r.HandleFunc("/v1/environments/{environment}", api.EnvironmentApiGet).Methods("GET")
	r.HandleFunc("/v1/environments/{environment}", api.EnvironmentApiUpdate).Methods("PUT")
	r.HandleFunc("/v1/environments/{environment}", api.EnvironmentApiDelete).Methods("DELETE")

	r.HandleFunc("/v1/watches", api.WatchApiList).Methods("GET")
	r.HandleFunc("/v1/watches", api.WatchApiCreate).Methods("POST")
	r.HandleFunc("/v1/watches/{watch}", api.WatchApiGet).Methods("GET")
	r.HandleFunc("/v1/watches/{watch}", api.WatchApiUpdate).Methods("PUT")
	r.HandleFunc("/v1/watches/{watch}", api.WatchApiDelete).Methods("DELETE")

	r.HandleFunc("/v1/triggers/time", api.TimeTriggerApiList).Methods("GET")
	r.HandleFunc("/v1/triggers/time", api.TimeTriggerApiCreate).Methods("POST")
	r.HandleFunc("/v1/triggers/time/{timeTrigger}", api.TimeTriggerApiGet).Methods("GET")
	r.HandleFunc("/v1/triggers/time/{timeTrigger}", api.TimeTriggerApiUpdate).Methods("PUT")
	r.HandleFunc("/v1/triggers/time/{timeTrigger}", api.TimeTriggerApiDelete).Methods("DELETE")

	r.HandleFunc("/proxy/{dbType}", api.FunctionLogsApiPost).Methods("POST")

	address := fmt.Sprintf(":%v", port)

	log.WithFields(log.Fields{"port": port}).Info("Server started")
	log.Fatal(http.ListenAndServe(address, handlers.LoggingHandler(os.Stdout, r)))
}
