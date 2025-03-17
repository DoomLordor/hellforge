package schemas

import (
	"errors"
	"fmt"
)

var TypeError = errors.New("schema not a struct")
var PointerError = errors.New("schema not pointer")
var FieldError = errors.New("schema field not implementation interface `Field`")
var ParseFieldError = errors.New("parse field error")
var RequiredError = errors.New("required error")
var SubStructError = errors.New("sub struct not a struct or pointer on struct")

func newParseError(name, value string) error {
	return fmt.Errorf(`%v: "%s" error from value "%s"`, ParseFieldError, name, value)
}

func newRequiredError(name string) error {
	return fmt.Errorf(`%v: "%s" required`, RequiredError, name)
}
