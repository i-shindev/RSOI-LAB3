package domain

func (in NewPaymentInput) Validate() (NewPayment, error) {
	var fields []FieldError
	switch {
	case in.Price == nil:
		fields = append(fields, FieldError{Field: "price", Message: "must be present"})
	case *in.Price < 0:
		fields = append(fields, FieldError{Field: "price", Message: "must not be negative"})
	}
	if len(fields) > 0 {
		return NewPayment{}, &ValidationError{Fields: fields}
	}
	return NewPayment{Price: *in.Price}, nil
}
