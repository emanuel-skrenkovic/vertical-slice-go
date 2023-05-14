package tql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type typeMapper struct {
	typeFieldCache map[string]map[string]int
}

var mapper = typeMapper{
	typeFieldCache: make(map[string]map[string]int),
}

type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func QuerySingleOrDefault[T any](ctx context.Context, q Querier, def T, query string, params ...any) (result T, err error) {
	result, err = QuerySingle[T](ctx, q, query, params)
	switch {
	case err != nil && errors.Is(err, sql.ErrNoRows):
		return def, nil
	case err != nil:
		return result, err
	default:
		return result, nil
	}
}

func QuerySingle[T any](ctx context.Context, q Querier, query string, params ...any) (result T, err error) {
	results, err := Query[T](ctx, q, query, params)
	if err != nil {
		return result, err
	}

	if len(results) > 1 {
		return result, fmt.Errorf("found more than one result")
	}

	return result, err
}

func QueryFirstOrDefault[T any](ctx context.Context, q Querier, def T, query string, params ...any) (result T, err error) {
	result, err = QueryFirst[T](ctx, q, query, params)
	switch {
	case err != nil && errors.Is(err, sql.ErrNoRows):
		return def, nil
	case err != nil:
		return result, err
	default:
		return result, nil
	}
}

func QueryFirst[T any](ctx context.Context, q Querier, query string, params ...any) (result T, err error) {
	rows, err := q.QueryContext(ctx, query, params...)
	if err != nil {
		return result, err
	}

	if rows == nil {
		return result, sql.ErrNoRows
	}

	defer func() {
		if rows.Err() != nil {
			return
		}

		if err = rows.Close(); err != nil {
			// #horribleways
			err = fmt.Errorf("failed to close rows: %w", err)
		}
	}()

	if !rows.Next() {
		return result, sql.ErrNoRows
	}

	val := reflect.Indirect(reflect.ValueOf(result))

	switch val.Kind() {
	case reflect.Struct:
		var cols []string
		cols, err = rows.Columns()
		if err != nil {
			return result, err
		}

		var dest []any
		dest, err = createDestinations(&result, cols)
		if err != nil {
			return result, err
		}

		if err = rows.Scan(dest...); err != nil {
			return result, err
		}

	case reflect.Slice:
		return result, fmt.Errorf("invalid type: slice")

	case reflect.Pointer:
		underlyingType := reflect.TypeOf(result).Elem()
		zero := reflect.New(underlyingType)

		val.Set(zero)

		if err = rows.Scan(result); err != nil {
			return result, err
		}

	default:
		if err = rows.Scan(&result); err != nil {
			return result, err
		}
	}

	return result, err
}

func Query[T any](ctx context.Context, q Querier, query string, params ...any) (result []T, err error) {
	result = make([]T, 0)

	rows, err := q.QueryContext(ctx, query, params...)
	if err != nil {
		return result, err
	}

	if rows == nil {
		return result, nil
	}

	defer func() {
		if rows.Err() != nil {
			return
		}

		if err = rows.Close(); err != nil {
			// #horribleways
			err = fmt.Errorf("failed to close rows: %w", err)
		}
	}()

	for rows.Next() {
		var current T

		val := reflect.Indirect(reflect.ValueOf(current))

		switch val.Kind() {
		case reflect.Struct:
			var cols []string
			cols, err = rows.Columns()
			if err != nil {
				return result, err
			}

			var dest []any
			dest, err = createDestinations(&current, cols)
			if err != nil {
				return result, err
			}

			if err = rows.Scan(dest...); err != nil {
				return result, err
			}

		case reflect.Pointer:
			underlyingType := reflect.TypeOf(current).Elem()
			zero := reflect.New(underlyingType)

			val.Set(zero)

			if err = rows.Scan(current); err != nil {
				return result, err
			}

		default:
			if err = rows.Scan(&current); err != nil {
				return result, err
			}
		}

		result = append(result, current)
	}

	return result, err
}

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func Exec(ctx context.Context, e Executor, query string, params ...any) (sql.Result, error) {
	parameters, err := mapParameters(params...)
	if err != nil {
		return nil, err
	}

	// #horribleways
	driverName := sql.Drivers()[0]
	pos, nam, err := parameterIndicators(driverName)
	if err != nil {
		return nil, err
	}

	parameterizedQuery, args, err := parameterizeQuery(pos, nam, query, parameters)
	if err != nil {
		return nil, err
	}

	if len(args) < 1 {
		args = params
	}

	return e.ExecContext(ctx, parameterizedQuery, args...)
}

func mapParameters(params ...any) (map[string]any, error) {
	parameters := make(map[string]any)

ParamLoop:
	for _, p := range params {
		if _, ok := p.(driver.Valuer); ok {
			continue
		}

		val := reflect.ValueOf(p)

		switch val.Kind() {
		case reflect.Map:
			value := reflect.Indirect(val).Interface()
			m := value.(map[string]any)

			for k, v := range m {
				if _, exists := parameters[k]; exists {
					return nil, fmt.Errorf("found parameter with duplicate name: %s", k)
				}

				parameters[k] = v
			}

		case reflect.Struct:
			// TODO: cache per type
			value := reflect.Indirect(val)

			valueType := reflect.TypeOf(p)
			typeName := valueType.Name()
			fieldsCount := valueType.NumField()

			fieldTags, found := typeFieldDbTags[typeName]
			if !found {
				fieldTags = make([]string, fieldsCount)
				typeFieldDbTags[typeName] = fieldTags

				for i := 0; i < fieldsCount; i++ {
					field := valueType.Field(i)

					if !field.IsExported() {
						continue ParamLoop
					}

					tag, found := field.Tag.Lookup("db")
					if !found {
						return nil, fmt.Errorf("field %s is not tagged with 'db' tag", field.Name)
					}

					fieldTags[i] = tag
				}
			}

			for i := 0; i < fieldsCount; i++ {
				// TODO: too inefficient!
				field := value.Field(i)
				sf := valueType.Field(i)
				if !sf.IsExported() {
					continue
				}
				parameters[fieldTags[i]] = field.Interface()
			}
		}
	}

	return parameters, nil
}

func isNameChar(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_'
}

type indicators struct {
	named      rune
	positional rune
}

var driverIndicators = map[string]indicators{
	"postgres":         {named: ':', positional: '$'},
	"pgx":              {named: ':', positional: '$'},
	"pq-timeouts":      {named: ':', positional: '$'},
	"cloudsqlpostgres": {named: ':', positional: '$'},
	"ql":               {named: ':', positional: '$'},
	"nrpostgres":       {named: ':', positional: '$'},
	"cockroach":        {named: ':', positional: '$'},

	"mysql":   {named: ':', positional: '?'},
	"nrmysql": {named: ':', positional: '?'},

	"sqlite3":   {named: ':', positional: '?'},
	"nrsqlite3": {named: ':', positional: '?'},
}

func parameterIndicators(driverName string) (rune, rune, error) {
	i, found := driverIndicators[driverName]
	if !found {
		return 0, 0, fmt.Errorf("failed to find driver parameter indicator mapping")
	}
	return i.named, i.positional, nil
}

func parameterizeQuery(
	namedParamIndicator rune,
	positionalParamIndicator rune,
	query string,
	parameters map[string]any,
) (string, []any, error) {
	var (
		insideName    bool
		hasPositional bool

		result     strings.Builder
		resultArgs = make([]any, 0, len(parameters))

		currentName strings.Builder
		currentNum  int
	)

	result.Grow(len(query))

	// TODO: inside name has to know the connection type to
	// properly decide on which token to use as the namedIndicator
	// of a parameter inside a query.
	// Also, which token to remap to.

	for _, c := range query {
		if !hasPositional && c == positionalParamIndicator {
			hasPositional = true
		}

		if !insideName && c == namedParamIndicator {
			currentName.Reset()
			insideName = true
			continue
		}

		if insideName && !isNameChar(c) {
			arg, found := parameters[currentName.String()]
			if !found {
				return "", []any{}, fmt.Errorf("query parameter '%s' not found in provided parameters", currentName.String())
			}
			resultArgs = append(resultArgs, arg)

			insideName = false
			currentNum++

			result.WriteRune(positionalParamIndicator)
			result.Write([]byte(strconv.Itoa(currentNum)))
			result.WriteRune(c)
			continue
		}

		if insideName {
			currentName.WriteRune(c)
			continue
		}

		result.WriteRune(c)
	}

	if hasPositional && len(resultArgs) > 0 {
		return "", []any{}, fmt.Errorf("mixed positional and named parameters")
	}

	return result.String(), resultArgs, nil
}

// TODO: candidate for caching
func createDestinations(source any, columns []string) ([]any, error) {
	value := reflect.ValueOf(source).Elem()
	valueType := value.Type()

	typeName := valueType.Name()
	if indices, found := mapper.typeFieldCache[typeName]; found {
		dest := make([]any, len(columns))
		for i, c := range columns {
			fieldIdx, found := indices[c]
			if !found {
				return nil, fmt.Errorf("no matching field found for column: %s", c)
			}

			field := value.Field(fieldIdx)
			if field.CanAddr() {
				dest[i] = field.Addr().Interface()
			} else {
				dest[i] = field.Interface()
			}
		}
		return dest, nil
	}

	numFields := valueType.NumField()
	indices := make(map[string]int, numFields)
	for i := 0; i < numFields; i++ {
		field := valueType.Field(i)

		tag, found := field.Tag.Lookup("db")
		if !found {
			continue
		}

		indices[tag] = i
	}

	dest := make([]any, len(columns))
	for i, c := range columns {
		fieldIdx, found := indices[c]
		if !found {
			return nil, fmt.Errorf("no matching field found for column: %s", c)
		}

		field := value.Field(fieldIdx)
		if field.CanAddr() {
			dest[i] = field.Addr().Interface()
		} else {
			dest[i] = field.Interface()
		}
	}

	mapper.typeFieldCache[typeName] = indices

	return dest, nil
}

// typeFieldDbTags
//
// Acts as a cache for struct field 'db' tag names.
// Used as such:
//
// tagName := typeFieldDbTags[typeName][fieldNumber]
var typeFieldDbTags = make(map[string][]string)

func bindArgs(params ...any) (map[string]any, error) {
	parameters := make(map[string]any)
	for _, p := range params {
		val := reflect.ValueOf(p)

		switch val.Kind() {
		case reflect.Map:
			value := reflect.Indirect(val).Interface()
			m := value.(map[string]any)

			for k, v := range m {
				if _, exists := parameters[k]; exists {
					return nil, fmt.Errorf("found parameter with duplicate name: %s", k)
				}

				parameters[k] = v
			}

		case reflect.Struct:
			// TODO: cache per type
			value := reflect.Indirect(val)
			valueType := reflect.TypeOf(p)

			// Aggressively pre-cache the struct 'db' tag bindings.
			typeName := valueType.Name()
			fieldsCount := valueType.NumField()

			fieldTags, found := typeFieldDbTags[typeName]
			if !found {
				fieldTags = make([]string, fieldsCount)
				typeFieldDbTags[typeName] = fieldTags

				for i := 0; i < fieldsCount; i++ {
					field := valueType.Field(i)
					tag, found := field.Tag.Lookup("db")
					if !found {
						return nil, fmt.Errorf("field %s is not tagged with 'db' tag", field.Name)
					}

					fieldTags[i] = tag
				}
			}

			for i := 0; i < fieldsCount; i++ {
				parameters[fieldTags[i]] = value.Field(i).Interface()
			}
		}
	}

	return parameters, nil
}
