package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/cards", app.listCardsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/cards", app.createCardHandler)
	router.HandlerFunc(http.MethodGet, "/v1/cards/:id", app.showCardHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/cards/:id", app.updateCardHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/cards/:id", app.deleteCardHandler)
	return router
}
