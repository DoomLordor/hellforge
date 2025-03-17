package schemas

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

const (
	tagSchema        = "schema"
	tagSchemaDefault = "default"
	tagRequired      = "required"
	tagSub           = "sub"
	tagSkip          = "-"
	emptyTagSchema   = ""
	emptyValue       = ""
	requiredTrue     = "true"
	subTrue          = "true"
	ValuesSplitter   = ","
)

type Field interface {
	Parse(raw, def string) error
	Set(set bool)
	IsMultiple() bool
}

var splitter = "."

func SetSplitter(newSplitter string) {
	splitter = newSplitter
}

func Parse(obj any, query url.Values, prefix string) (url.Values, error) {
	var err error

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	} else {
		return nil, PointerError
	}

	if v.Kind() != reflect.Struct {
		return nil, TypeError
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldValue := v.Field(i)

		fieldStruct := t.Field(i)
		name := fieldStruct.Tag.Get(tagSchema)
		def := fieldStruct.Tag.Get(tagSchemaDefault)
		required := fieldStruct.Tag.Get(tagRequired)
		sub := fieldStruct.Tag.Get(tagSub) == subTrue

		if name == tagSkip {
			continue
		}

		if prefix != "" {
			name = fmt.Sprintf(`%s%s%s`, prefix, splitter, name)
		}

		if fieldValue.Kind() == reflect.Pointer && fieldValue.IsNil() {
			newField := reflect.New(fieldValue.Type().Elem())
			fieldValue.Set(newField)
		}

		if name == emptyTagSchema || sub {

			if fieldValue.Kind() == reflect.Struct {
				fieldValue = fieldValue.Addr()
			}

			if fieldValue.Elem().Kind() == reflect.Struct {
				query, err = Parse(fieldValue.Interface(), query, name)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, SubStructError
			}

			if name == emptyTagSchema {
				continue
			}
		}

		if fieldValue.Kind() != reflect.Pointer {
			fieldValue = fieldValue.Addr()
		}

		interfaceValue := fieldValue.Interface()
		field, ok := interfaceValue.(Field)

		if !ok {
			if sub {
				continue
			}
			return nil, FieldError
		}

		values, ok := query[name]
		if !field.IsMultiple() && len(values) > 1 {
			values = values[:1:1]
		}

		val := strings.Join(values, ValuesSplitter)
		if val == emptyValue && required == requiredTrue {
			return nil, newRequiredError(name)
		}

		delete(query, name)

		field.Set(ok)
		err = field.Parse(val, def)
		if err != nil {
			return nil, newParseError(name, val)
		}
	}

	return query, nil
}
