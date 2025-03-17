package logger

import (
	"fmt"

	"github.com/rs/zerolog"
)

type Event struct {
	event *zerolog.Event
}

func (e *Event) Msg(msg string) {
	e.event.Msg(msg)
}

func (e *Event) Send() {
	e.event.Send()
}

func (e *Event) Msgf(format string, v ...interface{}) {
	e.event.Msg(fmt.Sprintf(format, v...))
}

func (e *Event) Str(key, val string) *Event {
	e.event.Str(key, val)
	return e
}

func (e *Event) Strs(key string, vals []string) *Event {
	e.event.Strs(key, vals)
	return e
}

func (e *Event) Err(err error) *Event {
	e.event.Err(err)
	return e
}

func (e *Event) Bool(key string, b bool) *Event {
	e.event.Bool(key, b)
	return e
}

func (e *Event) Int(key string, i int) *Event {
	e.event.Int(key, i)
	return e
}

func (e *Event) Int64(key string, i int64) *Event {
	e.event.Int64(key, i)
	return e
}

func (e *Event) Uint(key string, i uint) *Event {
	e.event.Uint(key, i)
	return e
}

func (e *Event) Uint64(key string, i uint64) *Event {
	e.event.Uint64(key, i)
	return e
}

func (e *Event) Float64(key string, f float64) *Event {
	e.event.Float64(key, f)
	return e
}
