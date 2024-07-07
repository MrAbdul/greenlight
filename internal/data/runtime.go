package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// a custom runtime type which is a type of int32
type Runtime int32

// implrmrny a marshaljson method on it so that it satisfies the json.marshler interface, this should return the json-encoded
// value for the movie runtime (<runtime> mins)
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r *Runtime) MarshalJSON() ([]byte, error) {
	//generate a string containg the movie runtime in the required format
	jsonValue := fmt.Sprintf("%d mins", r)
	//use the strconv.Quote function on the string to wrap it in double quotes. it needs to be surronded by double quotes in order to be a valid *JSON string*
	quotedJSONValue := strconv.Quote(jsonValue)
	//convert it to a byte slice and return it
	return []byte(quotedJSONValue), nil

}

func (r *Runtime) UnmarshalJSON(bytes []byte) error {
	// We expect that the incoming JSON value will be a string in the format
	// "<runtime> mins", and the first thing we need to do is remove the surrounding
	// double-quotes from this string. If we can't unquote it, then we return the
	// ErrInvalidRuntimeFormat error.
	unquotedJSONValue, err := strconv.Unquote(string(bytes))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number.
	parts := strings.Split(unquotedJSONValue, " ")

	// Sanity check the parts of the string to make sure it was in the expected format.
	// If it isn't, we return the ErrInvalidRuntimeFormat error again.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Otherwise, parse the string containing the number into an int32. Again, if this
	// fails return the ErrInvalidRuntimeFormat error.
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Convert the int32 to a Runtime type and assign this to the receiver. Note that we
	// use the * operator to deference the receiver (which is a pointer to a Runtime
	// type) in order to set the underlying value of the pointer.
	*r = Runtime(i)

	return nil
}
