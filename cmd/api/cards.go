package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/vynquoc/cs-flash-cards/internal/data"
	"github.com/vynquoc/cs-flash-cards/internal/validator"
)

const (
	Tomorrow    = 1
	ThreeDays   = 3
	OneWeek     = 7
	TwoWeeks    = 14
	OneMonth    = 30
	ThreeMonths = 90
)

func (app *application) createCardHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title          string            `json:"title"`
		Tags           []string          `json:"tags"`
		Content        string            `json:"content"`
		CodeSnippet    *data.CodeSnippet `json:"code_snippet"`
		NextReviewDate time.Time         `json:"next_review_date"`
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

	card.NextReviewDate = app.calculateReviewDate(time.Now().Truncate(24*time.Hour), 1)

	if input.CodeSnippet != nil {
		card.CodeSnippet = input.CodeSnippet
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

	card, err := app.models.Cards.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"card": card}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	card, err := app.models.Cards.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	var input struct {
		Title            *string           `json:"title"`
		Tags             []string          `json:"tags"`
		Content          *string           `json:"content"`
		CodeSnippet      *data.CodeSnippet `json:"code_snippet"`
		UpdateReviewDate *bool             `json:"update_review_date"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.Title != nil {
		card.Title = *input.Title
	}
	if input.Content != nil {
		card.Content = *input.Content
	}
	if input.Tags != nil {
		card.Tags = input.Tags
	}
	if input.CodeSnippet != nil {
		card.CodeSnippet = input.CodeSnippet
	}
	if input.UpdateReviewDate != nil && *input.UpdateReviewDate {
		var days int
		daysDifferent := int(card.NextReviewDate.Sub(card.CreatedAt).Hours() / 24)
		fmt.Println(daysDifferent)
		switch daysDifferent {
		case Tomorrow:
			days = ThreeDays
		case ThreeDays:
			days = OneWeek
		case OneWeek:
			days = TwoWeeks
		case TwoWeeks:
			days = OneMonth
		default:
			days = ThreeMonths
		}
		card.NextReviewDate = app.calculateReviewDate(card.CreatedAt, days)
	}
	v := validator.New()
	if data.ValidateCard(v, card); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Cards.Update(card)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"card": card}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.Cards.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "card successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listCardsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string
		Tags  []string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Tags = app.readCSV(qs, "tags", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "created_at")
	input.Filters.SortSafeList = []string{"id", "title", "created_at", "next_review_date", "-id", "-title", "-created_at", "-next_review_date"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	cards, metadata, err := app.models.Cards.GetAll(input.Title, input.Tags, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"cards": cards, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listReviewCardHandler(w http.ResponseWriter, r *http.Request) {
	cards, err := app.models.Cards.GetReviewCards()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"cards": cards}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showRandomCard(w http.ResponseWriter, r *http.Request) {
	card, err := app.models.Cards.GetRandomCard()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"card": card}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
