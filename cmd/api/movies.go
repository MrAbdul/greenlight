package main

import (
	"fmt"
	"greenlight.abdulalsh.com/internal/data"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	//an instance of the movie struct with the id we got from the url and some dummy data, notice that we don't set a value for the year

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}
	//we will encode the struct to json and send it as http response
	err = app.writeJSON(w, http.StatusOK, movie, nil)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The sevrer encounterd an error and couldn't process your request", http.StatusInternalServerError)
	}
	//otherwise, interpolate the movie id in a placeholder responose
}