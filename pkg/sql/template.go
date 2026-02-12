package sql

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// sqlEscape escapes single quotes in a string for SQL queries.
// It doubles any single quote characters to prevent SQL injection.
func sqlEscape(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

// sqlString wraps a string value in single quotes for SQL queries.
func sqlString(value string) string {
	return fmt.Sprintf("'%s'", value)
}

// sqlStringLike formats a string for use in SQL LIKE queries, wrapping the value with '%' and escaping it.
func sqlStringLike(value string) string {
	return fmt.Sprintf("'%%%s%%'", sqlEscape(value))
}

// sqlStringIn converts a slice of values into a SQL IN clause string, properly escaping and quoting each value.
// Returns an error if the input cannot be cast to a slice of strings.
func sqlStringIn(values interface{}) (string, error) {
	strList, err := cast.CastListToStringList(values)
	if err != nil {
		return "", errors.Errorf("could not cast %v to []string", values)
	}
	return fmt.Sprintf("'%s'", strings.Join(strList, "','")), nil
}

// sqlIn converts a slice of interface{} values into a comma-separated string for SQL queries.
// Each value is formatted using fmt.Sprintf with the %v verb.
func sqlIn(values []interface{}) string {
	strValues := make([]string, len(values))
	for i, v := range values {
		strValues[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(strValues, ",")
}

// sqlIntIn converts a slice of integer values into a comma-separated string for SQL queries.
// Returns an empty string if the input cannot be cast to a slice of int64.
func sqlIntIn(values interface{}) string {
	v_, ok := cast.CastInterfaceToIntList[int64](values)
	if !ok {
		return ""
	}
	strValues := make([]string, len(v_))
	for i, v := range v_ {
		strValues[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(strValues, ",")
}

// sqlDate_ formats a date value for SQL queries, using different formats based on the date's timezone.
// Returns an error if the date cannot be parsed or formatted.
// This is a helper function used by other date formatting functions.
func sqlDate_(date interface{}, fullFormat string, defaultFormat string) (string, error) {
	switch v := date.(type) {
	case string:
		parsedDate, err := fields.ParseDate(v)
		if err != nil {
			return "", err
		}
		// if timezone is local, output YYYY-mm-dd
		if parsedDate.Location() == time.Local {
			return "'" + parsedDate.Format(defaultFormat) + "'", nil
		}
		return "'" + parsedDate.Format(fullFormat) + "'", nil
	case time.Time:
		if v.Location() == time.Local {
			return "'" + v.Format(defaultFormat) + "'", nil
		}
		return "'" + v.Format(fullFormat) + "'", nil
	default:
		return "", errors.Errorf("could not parse date %v", date)
	}
}

// sqlDate formats a date value for SQL queries as YYYY-MM-DD or RFC3339, based on the date's timezone.
// Returns an error if the date cannot be parsed or formatted.
func sqlDate(date interface{}) (string, error) {
	return sqlDate_(date, time.RFC3339, "2006-01-02")
}

// sqlDateTime formats a datetime value for SQL queries as YYYY-MM-DDTHH:MM:SS or RFC3339, based on the datetime's timezone.
// Returns an error if the datetime cannot be parsed or formatted.
func sqlDateTime(date interface{}) (string, error) {
	return sqlDate_(date, time.RFC3339, "2006-01-02T15:04:05")
}

// sqliteDate formats a date value specifically for SQLite queries as YYYY-MM-DD.
// Returns an error if the date cannot be parsed or formatted.
func sqliteDate(date interface{}) (string, error) {
	return sqlDate_(date, "2006-01-02", "2006-01-02")
}

// sqliteDateTime formats a datetime value specifically for SQLite queries as YYYY-MM-DD HH:MM:SS.
// Returns an error if the datetime cannot be parsed or formatted.
func sqliteDateTime(date interface{}) (string, error) {
	return sqlDate_(date, "2006-01-02 15:04:05", "2006-01-02 15:04:05")
}

// sqlLike formats a string for use in SQL LIKE queries by wrapping the value with '%'.
func sqlLike(value string) string {
	return "'%" + value + "%'"
}

// TODO(manuel, 2023-11-19) Wrap this in a templating class that can accept additional funcmaps
// (and maybe more templating functionality)

func CreateTemplate(
	ctx context.Context,
	subQueries map[string]string,
	ps map[string]interface{},
	db *sqlx.DB,
) *template.Template {
	t2 := templating.CreateTemplate("query").
		Funcs(templating.TemplateFuncs).
		Funcs(template.FuncMap{
			"sqlStringIn":    sqlStringIn,
			"sqlStringLike":  sqlStringLike,
			"sqlIntIn":       sqlIntIn,
			"sqlIn":          sqlIn,
			"sqlDate":        sqlDate,
			"sqlDateTime":    sqlDateTime,
			"sqliteDate":     sqliteDate,
			"sqliteDateTime": sqliteDateTime,
			"sqlLike":        sqlLike,
			"sqlString":      sqlString,
			"sqlEscape":      sqlEscape,
			"subQuery": func(name string) (string, error) {
				s, ok := subQueries[name]
				if !ok {
					return "", errors.Errorf("Subquery %s not found", name)
				}
				return s, nil
			},
			"sqlSlice": func(query string, args ...interface{}) ([]interface{}, error) {
				data, err := mergeQueryData(ps, args)
				if err != nil {
					return nil, err
				}
				_, rows, err := RunQuery(ctx, subQueries, query, data, db)
				if err != nil {
					// TODO(manuel, 2023-03-27) This nesting of errors in nested templates becomes quite unpalatable
					// This is what can be output for just one level deep:
					//
					// Error: Could not generate query: template: query:1:13: executing "query" at <sqlColumn (subQuery "post_types")>: error calling sqlColumn: Could not run query: SELECT post_type
					// FROM wp_posts
					// GROUP BY post_type
					// ORDER BY post_type
					// : Error 1146 (42S02): Table 'ttc_analytics.wp_posts' doesn't exist
					// exit status 1
					//
					// Make better error messages:
					return nil, errors.Wrapf(err, "Could not run query: %s", query)
				}
				defer func(rows *sqlx.Rows) {
					_ = rows.Close()
				}(rows)

				ret := []interface{}{}

				for rows.Next() {
					ret_, err := rows.SliceScan()
					if err != nil {
						return nil, errors.Wrapf(err, "Could not scan query: %s", query)
					}

					row := make([]interface{}, len(ret_))
					for i, v := range ret_ {
						row[i] = sqlEltToTemplateValue(v)
					}

					ret = append(ret, row)
				}

				return ret, nil
			},
			"sqlColumn": func(query string, args ...interface{}) ([]interface{}, error) {
				data, err := mergeQueryData(ps, args)
				if err != nil {
					return nil, err
				}
				renderedQuery, rows, err := RunQuery(ctx, subQueries, query, data, db)
				if err != nil {
					return nil, errors.Wrapf(err, "Could not run query: %s", renderedQuery)
				}
				defer func(rows *sqlx.Rows) {
					_ = rows.Close()
				}(rows)

				ret := make([]interface{}, 0)
				for rows.Next() {
					rows_, err := rows.SliceScan()
					if err != nil {
						return nil, errors.Wrapf(err, "Could not scan query: %s", renderedQuery)
					}

					if len(rows_) != 1 {
						return nil, errors.Errorf("Expected 1 column, got %d", len(rows_))
					}
					elt := rows_[0]

					v := sqlEltToTemplateValue(elt)

					ret = append(ret, v)
				}

				return ret, nil
			},
			"sqlSingle": func(query string, args ...interface{}) (interface{}, error) {
				data, err := mergeQueryData(ps, args)
				if err != nil {
					return nil, err
				}
				renderedQuery, rows, err := RunQuery(ctx, subQueries, query, data, db)
				if err != nil {
					return nil, errors.Wrapf(err, "Could not run query: %s", renderedQuery)
				}
				defer func(rows *sqlx.Rows) {
					_ = rows.Close()
				}(rows)

				ret := make([]interface{}, 0)
				if rows.Next() {
					rows_, err := rows.SliceScan()
					if err != nil {
						return nil, errors.Wrapf(err, "Could not scan query: %s", renderedQuery)
					}

					if len(rows_) != 1 {
						return nil, errors.Errorf("Expected 1 column, got %d", len(rows_))
					}

					ret = append(ret, rows_[0])
				}

				if rows.Next() {
					return nil, errors.Errorf("Expected 1 row, got more")
				}

				if len(ret) == 0 {
					return nil, nil
				}

				if len(ret) > 1 {
					return nil, errors.Errorf("Expected 1 row, got %d", len(ret))
				}

				return sqlEltToTemplateValue(ret[0]), nil
			},
			"sqlMap": func(query string, args ...interface{}) (interface{}, error) {
				data, err := mergeQueryData(ps, args)
				if err != nil {
					return nil, err
				}
				renderedQuery, rows, err := RunQuery(ctx, subQueries, query, data, db)
				if err != nil {
					return nil, errors.Wrapf(err, "Could not run query: %s", renderedQuery)
				}
				defer func(rows *sqlx.Rows) {
					_ = rows.Close()
				}(rows)

				ret := []map[string]interface{}{}

				for rows.Next() {
					ret_ := make(map[string]interface{})
					err = rows.MapScan(ret_)
					if err != nil {
						return nil, errors.Wrapf(err, "Could not scan query: %s", renderedQuery)
					}

					row := make(map[string]interface{})
					for k, v := range ret_ {
						row[k] = sqlEltToTemplateValue(v)
					}

					ret = append(ret, row)
				}

				return ret, nil
			},
		})

	return t2
}

func sqlEltToTemplateValue(elt interface{}) interface{} {
	switch v := elt.(type) {
	case []byte:
		return string(v)
	default:
		return v
	}
}

func CleanQuery(query string) string {
	// remove all empty whitespace lines
	v := filter(
		strings.Split(query, "\n"),
		func(s string) bool {
			return strings.TrimSpace(s) != ""
		},
	)
	query = strings.Join(
		smap(v, func(s string) string {
			return strings.TrimRight(s, " \t")
		}),
		"\n",
	)

	return query
}

func smap(strs []string, f func(s string) string) []string {
	ret := make([]string, len(strs))
	for i, s := range strs {
		ret[i] = f(s)
	}
	return ret
}

func filter(strs []string, f func(s string) bool) []string {
	ret := make([]string, 0, len(strs))
	for _, s := range strs {
		if f(s) {
			ret = append(ret, s)
		}
	}
	return ret
}
