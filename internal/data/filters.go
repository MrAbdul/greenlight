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
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// The calculateMetadata() function calculates the appropriate pagination metadata
// values given the total number of records, current page, and page size values. Note
// that when the last page value is calculated we are dividing two int values, and
// when dividing integer types in Go the result will also be an integer type, with
// the modulus (or remainder) dropped. So, for example, if there were 12 records in total
// and a page size of 5, the last page value would be (12+5-1)/5 = 3.2, which is then
// truncated to 3 by Go.
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		// Note that we return an empty Metadata struct if there are no records.
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
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
func (f Filters) limit() int {
	return f.PageSize
}
func (f Filters) offset() int {
	//Note: In the offset() method there is the theoretical risk of an integer overflow as we are multiplying two int values together. However, this is mitigated by the validation rules we created in our ValidateFilters() function, where we enforced maximum values of page_size=100 and page=10000000 (10 million). This means that the value returned by offset() should never come close to overflowing.
	return (f.Page - 1) * f.PageSize
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
