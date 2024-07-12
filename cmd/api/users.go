package main

import (
	"errors"
	"greenlight.abdulalsh.com/internal/data"
	"greenlight.abdulalsh.com/internal/validator"
	"net/http"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the data from the request body into a new User struct. Notice also that we
	// set the Activated field to false, which isn't strictly necessary because the
	// Activated field will have the zero-value of false by default. But setting this
	// explicitly helps to make our intentions clear to anyone reading the code.
	//copy the data into a user struct
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	//insert data into db
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Launch a goroutine which runs an anonymous function that sends the welcome email.
	go func() {
		/*
			The code running in the background goroutine forms a closure over the user and app variables.
			It’s important to be aware that these ‘closed over’ variables are not scoped to the background goroutine,
			which means that any changes you make to them will be reflected in the rest of your codebase.
			For a simple example of this, see the following https://go.dev/play/p/eTz1xBm4W2a

			In our case we aren’t changing the value of these variables in any way, so this behavior won’t cause us any issues. But it is important to keep in mind.
		*/
		err = app.mailer.Send(user.Email, "user_welcome.gohtml", user)
		if err != nil {
			// Importantly, if there is an error sending the email then we use the
			// app.logger.Error() helper to manage it, instead of the
			// app.serverErrorResponse() helper like before.
			app.logger.Error(err.Error())
		}
	}()
	// Write a JSON response containing the user data along with a 201 Created status
	// code.
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
