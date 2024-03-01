package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/vynquoc/cs-flash-cards/internal/data"
	"github.com/vynquoc/cs-flash-cards/internal/validator"
)

func (app *application) createCardHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string            `json:"title"`
		Tags        []string          `json:"tags"`
		Content     string            `json:"content"`
		CodeSnippet *data.CodeSnippet `json:"code_snippet,omitempty"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	card := &data.Card{
		Title:   input.Title,
		Content: input.Content,
		Tags:    input.Tags,
	}

	if input.CodeSnippet != nil {
		card.CodeSnippet = *input.CodeSnippet
	}

	v := validator.New()
	if data.ValidateCard(v, card); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Cards.Insert(card)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/cards/%d", card.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"card": card}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	card := data.Card{
		ID:             id,
		CreatedAt:      time.Now(),
		Title:          "This is first card",
		Content:        "Long content",
		NextReviewDate: time.Now(),
		Tags:           []string{"Design Patterns, DSA"},
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"card": card}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
