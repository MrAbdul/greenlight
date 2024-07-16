package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"greenlight.abdulalsh.com/internal/validator"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
)

// Retrieve the "id" URL parameter from the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return 0 and an error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	//httprouter will store any interpolated url parameter in the request context. we can use the paramsfromcontext to retrive a slice containg these parameter names and values
	params := httprouter.ParamsFromContext(r.Context())

	//we convert the id parameter to an in of base 10 and 64 bits in size, if we can't or less than one we surve a notfound
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// we define an envelope type to better represent our data
type envelope map[string]any

// Define a writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one.
	//while marshal indent is better from a readability point of view, its more expensive on the system

	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	// At this point, we know that we won't encounter any more errors before writing the
	// response, so it's safe to add any headers that we want to include. We loop
	// through the header map and add each header to the http.ResponseWriter header map.
	// Note that it's OK if the provided header map is nil. Go doesn't throw an error
	// if you try to range over (or generally, read from) a nil map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header, then write the status code and
	// JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	//we will set a max byte
	maxBytes := 10_084_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	//init the json decoder and call the disallow unknown fields,
	dec := json.NewDecoder(r.Body)
	//it will raise an error if an unkwown field gets decoaded
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)

	if err != nil {
		// If there is an error during decoding, start the triage...
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		//max byte error var
		var maxByteError *http.MaxBytesError
		switch {
		// Use the errors.As() function to check whether the error has the type
		// *json.SyntaxError. If it does, then return a plain-english error message
		// which includes the location of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON. So we check for this using errors.Is() and
		// return a generic error message. There is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the
		// JSON value is the wrong type for the target destination. If the error relates
		// to a specific field, then we include that in our error message to make it
		// easier for the client to debug.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is empty. We
		// check for this with errors.Is() and return a plain-english error message
		// instead.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		//if the json contains a field which cannot be mapped to the target, we will get an error
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown field %q", fieldName)

		//if the error as saya that we have max bytes, then we exceeded the max size
		case errors.As(err, &maxByteError):
			return fmt.Errorf("body must not be larger than %d", maxByteError.Limit)

		// A json.InvalidUnmarshalError error will be returned if we pass something
		// that is not a non-nil pointer to Decode(). We catch this and panic,
		// rather than returning an error to our handler. At the end of this chapter
		// we'll talk about panicking versus returning errors, and discuss why it's an
		// appropriate thing to do in this specific situation.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// For anything else, return the error message as-is.
		default:
			return err
		}
	}
	// Call Decode() again, using a pointer to an empty anonymous struct as the
	// destination. If the request body only contained a single JSON value this will
	// return an io.EOF error. So if we get anything else, we know that there is
	// additional data in the request body and we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	istring := qs.Get(key)
	if istring == "" {
		return defaultValue
	}
	if i, err := strconv.Atoi(istring); err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	} else {
		return i
	}
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	csv := strings.Split(s, ",")
	return csv
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, map[string]any{"Error": message})

}

// The background() helper accepts an arbitrary function as a parameter.
// This background() helper leverages the fact that Go has first-class functions,
// which means that functions can be assigned to variables and passed as parameters to other functions.
func (app *application) background(fn func()) {
	//increment the waitgroup counter
	app.wg.Add(1)
	// Launch a background goroutine.
	go func() {
		//defer to decrenebt the waitgroup counter before the goroutine returns
		defer app.wg.Done()
		// Recover any panic.
		/*
			The code running in the background goroutine forms a closure over the user and app variables.
			It’s important to be aware that these ‘closed over’ variables are not scoped to the background goroutine,
			which means that any changes you make to them will be reflected in the rest of your codebase.
			For a simple example of this, see the following https://go.dev/play/p/eTz1xBm4W2a

			In our case we aren’t changing the value of these variables in any way, so this behavior won’t cause us any issues. But it is important to keep in mind.
		*/
		//****Recovering panic****//
		// Run a deferred function which uses recover() to catch any panic, and log an
		// error message instead of terminating the application.
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		// Execute the arbitrary function that we passed as the parameter.
		fn()
	}()
}

// The serverError helper writes a log entry at Error level (including the request
// method and URI as attributes), then sends a generic 500 Internal Server Error
// response to the user.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	// Retrieve the appropriate template set from the cache based on the page
	// name (like 'home.gohtml'). If no entry exists in the cache with the
	// provided name, then create a new error and call the serverError() helper
	// method that we made earlier and return.
	template, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	//to handle errors, we will first write the content to a buffer and if thats ok, we will stream it to the user,
	// if not we will render an error
	buf := new(bytes.Buffer)
	// we will write the template to the buffer, instead of straight to hte response writer, if htere is an error we will call our server error
	err := template.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	////we introduce a delibrate error to see the server error log in action
	////we saw that we can set a custome slog to the server struct to normalize error logging
	//w.Header().Set("Content-Length", "this is a delibrate error")
	//if the template is written to the buffer without any errors, we are safe to go ahead and write the status code.
	w.WriteHeader(status)

	//then we write the contents of the buffer to the writer.
	buf.WriteTo(w)

}
