package main

import (
	"errors"
	"fmt"
	"greenlight.abdulalsh.com/internal/data"
	"greenlight.abdulalsh.com/internal/validator"
	"net/http"
)

// getCategoryHandler
func (app *application) getCategoryHandler(w http.ResponseWriter, r *http.Request) {
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
	category, err := app.models.CategoryModel.Get(id, lang)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"category": category}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getCategoriesHandler
func (app *application) getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	lang, v := app.readLanguageHeader(r)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	categories, err := app.models.CategoryModel.GetAll(lang)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"category": categories}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// createCategoryHandler
func (app *application) createCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		Image string `json:"image"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorErrResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}
	category := &data.Category{
		Title:    input.Title,
		Language: "en",
		Image:    input.Image,
	}
	v := validator.New()
	data.ValidateCategory(v, category)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.CategoryModel.Insert(category)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("%s/%d", categoriesV1, category.ID))
	//write the json response with a 201 created status code the movie data in the response body and the location header
	err = app.writeJSON(w, http.StatusCreated, envelope{"category": category}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCategoryLanugageHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Id       int64  `json:"id"`
		Title    string `json:"title"`
		Language string `json:"language"`
		Image    string `json:"image"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorErrResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}
	//get the cat to set default image if image doesn't exists.
	category, err := app.models.CategoryModel.Get(input.Id, "en")
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
	category.Language = input.Language
	category.Title = input.Title
	if input.Image != "" {
		category.Image = input.Image
	}

	v := validator.New()
	data.ValidateCategory(v, category)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.CategoryModel.Update(category)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, data.ErrDublicateCategoryTranslation):
			v.AddError("category", "duplicate category translation please update category")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("%s/%d", categoriesV1, category.ID))
	//write the json response with a 201 created status code the movie data in the response body and the location header
	err = app.writeJSON(w, http.StatusCreated, envelope{"category": category}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// deleteCategoryHandler
func (app *application) deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.CategoryModel.Delete(id)

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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Category deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
