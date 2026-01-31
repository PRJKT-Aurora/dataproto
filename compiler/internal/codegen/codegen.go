// Package codegen provides code generation from DataProto AST.
package codegen

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/aurora/dataproto/internal/parser"
)

// Generator is the interface for all code generators.
type Generator interface {
	Generate(file *parser.File) (map[string]string, error)
}

// TypeMapping provides type mappings for different backends.
type TypeMapping struct {
	Proto    string
	SQLite   string
	Postgres string
	Java     string
	Swift    string
	Python   string
}

// GetTypeMapping returns the type mapping for a DataProto type.
func GetTypeMapping(typeName string) TypeMapping {
	switch typeName {
	case "string":
		return TypeMapping{
			Proto:    "string",
			SQLite:   "TEXT",
			Postgres: "TEXT",
			Java:     "String",
			Swift:    "String",
			Python:   "str",
		}
	case "int32":
		return TypeMapping{
			Proto:    "int32",
			SQLite:   "INTEGER",
			Postgres: "INTEGER",
			Java:     "int",
			Swift:    "Int32",
			Python:   "int",
		}
	case "int64":
		return TypeMapping{
			Proto:    "int64",
			SQLite:   "INTEGER",
			Postgres: "BIGINT",
			Java:     "long",
			Swift:    "Int64",
			Python:   "int",
		}
	case "float":
		return TypeMapping{
			Proto:    "float",
			SQLite:   "REAL",
			Postgres: "REAL",
			Java:     "float",
			Swift:    "Float",
			Python:   "float",
		}
	case "double":
		return TypeMapping{
			Proto:    "double",
			SQLite:   "REAL",
			Postgres: "DOUBLE PRECISION",
			Java:     "double",
			Swift:    "Double",
			Python:   "float",
		}
	case "bool":
		return TypeMapping{
			Proto:    "bool",
			SQLite:   "INTEGER",
			Postgres: "BOOLEAN",
			Java:     "boolean",
			Swift:    "Bool",
			Python:   "bool",
		}
	case "bytes":
		return TypeMapping{
			Proto:    "bytes",
			SQLite:   "BLOB",
			Postgres: "BYTEA",
			Java:     "byte[]",
			Swift:    "Data",
			Python:   "bytes",
		}
	case "timestamp":
		return TypeMapping{
			Proto:    "int64",
			SQLite:   "INTEGER",
			Postgres: "BIGINT",
			Java:     "long",
			Swift:    "Int64",
			Python:   "int",
		}
	default:
		// Custom type (enum or entity reference)
		return TypeMapping{
			Proto:    typeName,
			SQLite:   "TEXT",
			Postgres: "TEXT",
			Java:     typeName,
			Swift:    typeName,
			Python:   typeName,
		}
	}
}

// ToPascalCase converts a string to PascalCase.
func ToPascalCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

// ToCamelCase converts a string to camelCase.
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}
	return strings.ToLower(string(pascal[0])) + pascal[1:]
}

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ToScreamingSnakeCase converts a string to SCREAMING_SNAKE_CASE.
func ToScreamingSnakeCase(s string) string {
	return strings.ToUpper(ToSnakeCase(s))
}

// splitWords splits a string into words based on underscores and case changes.
func splitWords(s string) []string {
	var words []string
	var current strings.Builder

	for i, r := range s {
		if r == '_' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			continue
		}

		if unicode.IsUpper(r) && i > 0 {
			prev := rune(s[i-1])
			if !unicode.IsUpper(prev) && prev != '_' {
				if current.Len() > 0 {
					words = append(words, current.String())
					current.Reset()
				}
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// ExprToSQL converts an expression AST to SQL string.
func ExprToSQL(expr parser.Expr) string {
	switch e := expr.(type) {
	case *parser.BinaryExpr:
		left := ExprToSQL(e.Left)
		right := ExprToSQL(e.Right)
		return fmt.Sprintf("%s %s %s", left, e.Op, right)

	case *parser.UnaryExpr:
		operand := ExprToSQL(e.Operand)
		return fmt.Sprintf("%s %s", e.Op, operand)

	case *parser.IsNullExpr:
		operand := ExprToSQL(e.Operand)
		if e.Not {
			return fmt.Sprintf("%s IS NOT NULL", operand)
		}
		return fmt.Sprintf("%s IS NULL", operand)

	case *parser.IdentExpr:
		return e.Name

	case *parser.LiteralExpr:
		switch v := e.Value.(type) {
		case string:
			return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		case int64:
			return fmt.Sprintf("%d", v)
		case float64:
			return fmt.Sprintf("%f", v)
		case bool:
			if v {
				return "1"
			}
			return "0"
		default:
			return "NULL"
		}

	case *parser.CallExpr:
		var args []string
		for _, arg := range e.Args {
			args = append(args, ExprToSQL(arg))
		}
		return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))

	case *parser.ParenExpr:
		return fmt.Sprintf("(%s)", ExprToSQL(e.Inner))

	default:
		return ""
	}
}

// ExprToSQLWithParams converts an expression to parameterized SQL.
// Returns the SQL string and a list of parameter names.
func ExprToSQLWithParams(expr parser.Expr, paramPrefix string) (string, []string) {
	var params []string
	sql := exprToSQLWithParamsInternal(expr, paramPrefix, &params)
	return sql, params
}

func exprToSQLWithParamsInternal(expr parser.Expr, prefix string, params *[]string) string {
	switch e := expr.(type) {
	case *parser.BinaryExpr:
		left := exprToSQLWithParamsInternal(e.Left, prefix, params)
		right := exprToSQLWithParamsInternal(e.Right, prefix, params)
		return fmt.Sprintf("%s %s %s", left, e.Op, right)

	case *parser.UnaryExpr:
		operand := exprToSQLWithParamsInternal(e.Operand, prefix, params)
		return fmt.Sprintf("%s %s", e.Op, operand)

	case *parser.IsNullExpr:
		operand := exprToSQLWithParamsInternal(e.Operand, prefix, params)
		if e.Not {
			return fmt.Sprintf("%s IS NOT NULL", operand)
		}
		return fmt.Sprintf("%s IS NULL", operand)

	case *parser.IdentExpr:
		// Check if this is a parameter reference (lowercase, matches query param)
		if isLowerCamelCase(e.Name) {
			*params = append(*params, e.Name)
			return "?"
		}
		return e.Name

	case *parser.LiteralExpr:
		switch v := e.Value.(type) {
		case string:
			return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		case int64:
			return fmt.Sprintf("%d", v)
		case float64:
			return fmt.Sprintf("%f", v)
		case bool:
			if v {
				return "1"
			}
			return "0"
		default:
			return "NULL"
		}

	case *parser.CallExpr:
		var args []string
		for _, arg := range e.Args {
			args = append(args, exprToSQLWithParamsInternal(arg, prefix, params))
		}
		// Handle special functions
		if e.Name == "NOW" {
			return "(strftime('%s', 'now') * 1000)"
		}
		return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))

	case *parser.ParenExpr:
		return fmt.Sprintf("(%s)", exprToSQLWithParamsInternal(e.Inner, prefix, params))

	default:
		return ""
	}
}

func isLowerCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	return unicode.IsLower(rune(s[0]))
}

// IndentLines indents each line of a string.
func IndentLines(s string, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if len(line) > 0 {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}
