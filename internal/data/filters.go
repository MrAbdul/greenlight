package data

import (
	"greenlight.abdulalsh.com/internal/validator"
	"strings"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func (f Filters) sortColumn() string {
	//even though this should have already been checked by validateFilters functions,
	//its sensible failsafe to help stop SQL injection attacks in case something passes the vaildation
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter " + f.Sort)
}
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}
func ValidateFilters(v *validator.Validator, f Filters) {
	//check the page, and page_size params contain sensible values
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "Must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	//check that the sort parameter matches a value in the safelist
	v.Check(validator.PermittedValues(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
