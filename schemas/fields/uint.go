package fields

import (
	"strconv"
)

type UintFiled struct {
	value uint64
	set   bool
}

func (f *UintFiled) Parse(raw, def string) error {
	f.value = 0

	if raw == emptyValue {
		raw = def
	}

	if raw == emptyValue {
		return nil
	}

	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return err
	}

	f.value = value
	return nil
}

func (f *UintFiled) Set(set bool) {
	f.set = set
}

func (f *UintFiled) IsSet() bool {
	return f.set
}

func (f *UintFiled) Value() uint64 {
	return f.value
}

func (f *UintFiled) IsMultiple() bool {
	return false
}
