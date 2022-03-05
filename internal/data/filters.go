package data

import (
	"github.com/DataDavD/snippetbox/greenlight/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

// ValidateFilters runs validation checks on the Filters type.
func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that page and page_size parameters contain sensible values.
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page <= 10_000_0000, "", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// Check that the sort parameter matches a value in the safelist.
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}
