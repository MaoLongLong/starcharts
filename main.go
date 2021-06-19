package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/maolonglong/starcharts/config"
	"github.com/maolonglong/starcharts/controller"
	"github.com/maolonglong/starcharts/internal/github"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

func main() {
	var log = log.WithField("port", config.Port())
	var cache = cache.New(5*time.Minute, 10*time.Minute)
	var github = github.New(cache)

	var r = mux.NewRouter()
	r.Path("/").
		Methods(http.MethodGet).
		HandlerFunc(controller.Index())
	r.PathPrefix("/static/").
		Methods(http.MethodGet).
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.Path("/{owner}/{repo}.svg").
		Methods(http.MethodGet).
		HandlerFunc(controller.GetRepoChart(github))
	r.Path("/{owner}/{repo}").
		Methods(http.MethodGet).
		HandlerFunc(controller.GetRepo(github))

	var srv = &http.Server{
		Handler:      r,
		Addr:         ":" + strconv.Itoa(config.Port()),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	log.Info("starting up...")
	log.WithError(srv.ListenAndServe()).Error("failed to start up server")
}
