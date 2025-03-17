package fields

import (
	"strconv"
)

type FloatFiled struct {
	value float64
	set   bool
}

func (f *FloatFiled) Parse(raw, def string) error {
	f.value = 0.0

	if raw == emptyValue {
		raw = def
	}

	if raw == emptyValue {
		return nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return err
	}

	f.value = value
	return nil
}

func (f *FloatFiled) Set(set bool) {
	f.set = set
}

func (f *FloatFiled) IsSet() bool {
	return f.set
}

func (f *FloatFiled) Value() float64 {
	return f.value
}

func (f *FloatFiled) IsMultiple() bool {
	return false
}
