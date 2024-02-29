package main

import (
	"fmt"
	"net/http"
)

func (app *application) createCardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Create cards")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}
