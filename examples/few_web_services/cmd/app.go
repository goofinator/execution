// Copyright (C) 2022 Blinnikov AA <goofinator@mail.ru>
//
// This file is part of execution.
//
// execution is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// execution is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with execution.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"net/http"

	"github.com/goofinator/execution"
	"github.com/goofinator/execution/examples/few_web_services/web_server"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	apiSrv := web_server.NewWebServer("api web server", 8080, NewApiHandler())
	monitoringSrv := web_server.NewWebServer("monitoring web server", 5000, NewStateHandler())

	config := execution.ExecutionManagerConfig{
		Log: log.StandardLogger(),
	}
	executionManager := execution.NewExecutionManager(config)
	executionManager.TakeOnControll(apiSrv, monitoringSrv)
	executionManager.ExecuteProcesses()
}

func NewApiHandler() *mux.Router {
	var cntr int
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/wisdom", func(w http.ResponseWriter, r *http.Request) {
		log.Info("GET: /api/v1/wisdom")
		fmt.Fprintln(w, "life is short")
	}).Methods("GET")

	r.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		cntr++
		if cntr > 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Info("GET: /api/v1/ping")
	}).Methods("GET")
	return r
}

func NewStateHandler() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		result, err := http.Get("http://localhost:8080/api/v1/ping")
		if err != nil || result.StatusCode >= 300 {
			log.Error("health check failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "looks healthy")
	}).Methods("GET")
	return r
}
