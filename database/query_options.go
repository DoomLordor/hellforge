package database

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
)

type SelectOption func(qb *goqu.SelectDataset) *goqu.SelectDataset

func ApplyOptions[R any](qb *goqu.SelectDataset, r R, opts func(R) []SelectOption) *goqu.SelectDataset {
	if r == nil {
		return qb
	}

	for _, opt := range opts(r) {
		qb = opt(qb)
	}

	return qb
}

func SliceFilter[T any](data []T, fieldName string, alias string) SelectOption {
	fieldName = GetFieldName(fieldName, alias)
	return func(qb *goqu.SelectDataset) *goqu.SelectDataset {
		if len(data) == 0 {
			return qb
		}

		return qb.Where(goqu.Ex{fieldName: data})
	}
}

func FieldFilter[T any](value T, fieldName, alias string) SelectOption {
	fieldName = GetFieldName(fieldName, alias)
	return func(qb *goqu.SelectDataset) *goqu.SelectDataset {
		return qb.Where(goqu.Ex{fieldName: value})
	}
}

func OptionalFieldFilter[T any](value *T, fieldName, alias string) SelectOption {
	fieldName = GetFieldName(fieldName, alias)
	return func(qb *goqu.SelectDataset) *goqu.SelectDataset {
		if value == nil {
			return qb
		}

		return qb.Where(goqu.Ex{fieldName: *value})
	}
}

func OptionalFieldFilterWithConverter[T, C any](value *T, fieldName, alias string, converter func(T) C) SelectOption {
	fieldName = GetFieldName(fieldName, alias)
	return func(qb *goqu.SelectDataset) *goqu.SelectDataset {
		if value == nil {
			return qb
		}

		return qb.Where(goqu.Ex{fieldName: converter(*value)})
	}
}

func GetFieldName(fieldName, alias string) string {
	if alias == "" {
		return fieldName
	}

	return fmt.Sprintf("%s.%s", alias, fieldName)
}
