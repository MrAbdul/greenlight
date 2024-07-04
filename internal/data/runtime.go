package data

import (
	"fmt"
	"strconv"
)

// a custom runtime type which is a type of int32
type Runtime int32

//implrmrny a marshaljson method on it so that it satisfies the json.marshler interface, this should return the json-encoded
//value for the movie runtime (<runtime> mins)

func (r Runtime) MarshalJSON() ([]byte, error) {
	//generate a string containg the movie runtime in the required format
	jsonValue := fmt.Sprintf("%d mins", r)
	//use the strconv.Quote function on the string to wrap it in double quotes. it needs to be surronded by double quotes in order to be a valid *JSON string*
	quotedJSONValue := strconv.Quote(jsonValue)
	//convert it to a byte slice and return it
	return []byte(quotedJSONValue), nil

}
