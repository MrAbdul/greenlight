package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	//init a new httprouter
	router := httprouter.New()
	//we convert the notFoundResponse() helper to a http.Handler using the http.HandlerFunc() adapter.
	//and we set it as the custom error handler for 404
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	//likewise for 405
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	//we register our routes here
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	return app.recoverPanic(app.rateLimit(router))
}
