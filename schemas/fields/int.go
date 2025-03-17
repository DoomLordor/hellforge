package fields

import (
	"strconv"
)

type IntFiled struct {
	value int64
	set   bool
}

func (f *IntFiled) Parse(raw, def string) error {
	f.value = 0

	if raw == emptyValue {
		raw = def
	}

	if raw == emptyValue {
		return nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}

	f.value = value
	return nil
}

func (f *IntFiled) Set(set bool) {
	f.set = set
}

func (f *IntFiled) IsSet() bool {
	return f.set
}

func (f *IntFiled) Value() int64 {
	return f.value
}

func (f *IntFiled) IsMultiple() bool {
	return false
}
