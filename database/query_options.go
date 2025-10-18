package database

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type SelectOption func(qb *goqu.SelectDataset) *goqu.SelectDataset

func ApplyOptions[R any](qb *goqu.SelectDataset, r *R, opts func(*R) []SelectOption) *goqu.SelectDataset {
	if r == nil {
		return qb
	}

	for _, opt := range opts(r) {
		qb = opt(qb)
	}

	return qb
}

func SliceFilter[T any](data []T, fieldName string, alias string) SelectOption {
	column := GetColumn(fieldName, alias)
	return func(qb *goqu.SelectDataset) *goqu.SelectDataset {
		if len(data) == 0 {
			return qb
		}

		return qb.Where(column.Eq(data))
	}
}

func FieldFilter[T any](value T, fieldName, alias string) SelectOption {
	column := GetColumn(fieldName, alias)
	return func(qb *goqu.SelectDataset) *goqu.SelectDataset {
		return qb.Where(column.Eq(value))
	}
}

func OptionalFieldFilter[T any](value *T, fieldName, alias string) SelectOption {
	column := GetColumn(fieldName, alias)
	return func(qb *goqu.SelectDataset) *goqu.SelectDataset {
		if value == nil {
			return qb
		}

		return qb.Where(column.Eq(*value))
	}
}

func OptionalFieldFilterWithConverter[T, C any](value *T, fieldName, alias string, converter func(T) C) SelectOption {
	column := GetColumn(fieldName, alias)
	return func(qb *goqu.SelectDataset) *goqu.SelectDataset {
		if value == nil {
			return qb
		}

		return qb.Where(column.Eq(converter(*value)))
	}
}

func GetColumn(fieldName, alias string) exp.IdentifierExpression {
	column := goqu.C(fieldName)
	if alias != "" {
		column = column.Table(alias)
	}

	return column
}
