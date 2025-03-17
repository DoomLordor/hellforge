package fields

import (
	"strings"
)

type StringsFiled struct {
	value []string
	set   bool
}

func (f *StringsFiled) Parse(raw, def string) error {
	f.value = []string{}

	if raw == emptyValue {
		raw = def
	}

	raw = strings.ReplaceAll(raw, ",,", ",")

	f.value = strings.Split(raw, ",")

	return nil
}

func (f *StringsFiled) Set(set bool) {
	f.set = set
}

func (f *StringsFiled) IsSet() bool {
	return f.set
}

func (f *StringsFiled) Value() []string {
	return f.value
}

func (f *StringsFiled) IsMultiple() bool {
	return true
}
