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

package web_server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				log.Errorf("panic during request: %s", r.URL)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func NewWebServer(name string, port int, handler http.Handler) WebServer {
	return WebServer{
		Server: &http.Server{
			Handler: Recovery(handler),
			Addr:    fmt.Sprintf(":%d", port),
		},
		name: name,
	}
}

type WebServer struct {
	*http.Server
	name string
}

func (r WebServer) Run(ctx context.Context) {
	go func() {
		err := r.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("%s ListenAndServe failure: %s", r.Name(), err)
		}
	}()
	<-ctx.Done()
	localCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := r.Shutdown(localCtx)
	if err != nil {
		log.Errorf("%s server shotdown failure: %s", r.Name(), err)
	}
}

func (r WebServer) Name() string {
	return r.name
}
