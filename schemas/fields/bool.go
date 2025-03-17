package fields

import (
	"strconv"
)

type BoolFiled struct {
	value bool
	set   bool
}

func (f *BoolFiled) Parse(raw, def string) error {
	f.value = false

	if raw == emptyValue {
		raw = def
	}

	if raw == emptyValue {
		return nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return err
	}

	f.value = value
	return nil
}

func (f *BoolFiled) Set(set bool) {
	f.set = set
}

func (f *BoolFiled) IsSet() bool {
	return f.set
}

func (f *BoolFiled) Value() bool {
	return f.value
}

func (f *BoolFiled) IsMultiple() bool {
	return false
}
