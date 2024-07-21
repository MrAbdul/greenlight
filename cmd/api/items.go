package main

import (
	"errors"
	"fmt"
	"greenlight.abdulalsh.com/internal/data"
	"greenlight.abdulalsh.com/internal/validator"
	"net/http"
)

// createItemHandler
func (app *application) createItemHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CategoryID int64  `json:"category_id"`
		Name       string `json:"name"`
		Image      string `json:"image"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorErrResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}
	item := &data.Item{
		CategoryID: input.CategoryID,
		Name:       input.Name,
		Image:      input.Image,
		Language:   "en",
	}
	v := validator.New()
	data.ValidateItem(v, item)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.ItemModel.Insert(item)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("%s/%d", itemsV1, item.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"item": item}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getItemHandler
func (app *application) getItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	lang, v := app.readLanguageHeader(r)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	item, err := app.models.ItemModel.Get(id, lang)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"item": item}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getItemsHandler
func (app *application) getItemsHandler(w http.ResponseWriter, r *http.Request) {
	lang, v := app.readLanguageHeader(r)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	items, err := app.models.ItemModel.GetAll(lang)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"items": items}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateItemHandler
func (app *application) updateItemHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID         int64  `json:"id"`
		CategoryID int64  `json:"category_id"`
		Name       string `json:"name"`
		Image      string `json:"image"`
		Language   string `json:"language"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorErrResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	item, err := app.models.ItemModel.Get(input.ID, input.Language)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	item.CategoryID = input.CategoryID
	item.Name = input.Name
	item.Image = input.Image
	item.Language = input.Language

	v := validator.New()
	data.ValidateItem(v, item)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.ItemModel.UpsertTranslation(item)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateItemTranslation):
			v.AddError("item", "duplicate item translation, please update item")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("%s/%d", itemsV1, item.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"item": item}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteItemHandler
func (app *application) deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.ItemModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Item deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
