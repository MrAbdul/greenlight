package main

import (
	"github.com/julienschmidt/httprouter"
	"greenlight.abdulalsh.com/internal/data"
	"greenlight.abdulalsh.com/ui"
	"net/http"
)

func (app *application) routes() http.Handler {
	//init a new httprouter
	// now we use mux.handle function to register the file server as handler for all url paths that start with /static/
	// for matching paths, we strip the "/static" prefix before the request reaches the file server
	//we don't need the session handling middleware for the static files
	router := httprouter.New()
	//we convert the notFoundResponse() helper to a http.Handler using the http.HandlerFunc() adapter.
	//and we set it as the custom error handler for 404
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	//likewise for 405
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	//we register our routes here
	//static files
	router.ServeFiles("/ui/*filepath", http.FS(ui.Files))
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission(data.MoviesRead, app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission(data.MoviesWrite, app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission(data.MoviesRead, app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission(data.MoviesWrite, app.deleteMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission(data.MoviesWrite, app.listMoviesHandler))

	//users
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.Handler(http.MethodPost, "/v1/users/activated", noSurf(http.HandlerFunc(app.activateUserHandler)))
	router.Handler(http.MethodGet, "/v1/users/activate", noSurf(http.HandlerFunc(app.activateUserFormGetHandler)))
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
