package main

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
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
