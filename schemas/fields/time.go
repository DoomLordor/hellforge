package fields

import (
	"time"
)

const (
	timeNow     = "now"
	layout      = "2006-01-02T15:04:05Z"
	layoutMilli = "2006-01-02T15:04:05.000Z"
)

type TimeFiled struct {
	value *time.Time
	set   bool
}

func (f *TimeFiled) Parse(raw, def string) error {
	f.value = nil

	if raw == emptyValue {
		if def == timeNow {
			now := time.Now().Truncate(time.Microsecond)
			f.value = &now
			return nil
		}
		raw = def
	}

	if raw == emptyValue {
		return nil
	}

	for _, format := range []string{layout, layoutMilli} {
		t, err := time.Parse(format, raw)
		if err == nil {
			f.value = &t
			return nil
		}
	}

	return nil
}

func (f *TimeFiled) Set(set bool) {
	f.set = set
}

func (f *TimeFiled) IsSet() bool {
	return f.set
}

func (f *TimeFiled) SetValue(value *time.Time) {
	f.value = value
}

func (f *TimeFiled) Value() *time.Time {
	return f.value
}

func (f *TimeFiled) IsMultiple() bool {
	return false
}
