package domain

import "strconv"

func (in ListCarsInput) Validate() (ListCarsQuery, error) {
	query := ListCarsQuery{Page: DefaultPage, Size: DefaultPageSize, ShowAll: in.ShowAll}
	var fields []FieldError
	if in.Page != nil {
		if *in.Page < 1 {
			fields = append(fields, FieldError{Field: "page", Message: "must be at least 1"})
		}
		query.Page = *in.Page
	}
	if in.Size != nil {
		if *in.Size < 1 || *in.Size > MaxPageSize {
			fields = append(fields, FieldError{
				Field:   "size",
				Message: "must be between 1 and " + strconv.Itoa(MaxPageSize),
			})
		}
		query.Size = *in.Size
	}
	if len(fields) > 0 {
		return ListCarsQuery{}, &ValidationError{Fields: fields}
	}
	return query, nil
}
