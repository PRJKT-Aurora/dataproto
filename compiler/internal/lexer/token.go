// Package lexer provides tokenization for DataProto schema files.
package lexer

// TokenType represents the type of a token.
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	IDENT     // identifier
	INT       // integer literal
	FLOAT     // float literal
	STRING    // string literal

	// Operators and delimiters
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	LBRACKET  // [
	RBRACKET  // ]
	SEMICOLON // ;
	COLON     // :
	COMMA     // ,
	DOT       // .
	AT        // @
	QUESTION  // ?
	EQUALS    // =
	BANG_EQ   // !=
	LT        // <
	LT_EQ     // <=
	GT        // >
	GT_EQ     // >=
	PLUS      // +
	MINUS     // -
	STAR      // *
	SLASH     // /
	PERCENT   // %
	CONCAT    // ||

	// Keywords
	PACKAGE
	IMPORT
	OPTION
	ENUM
	ENTITY
	QUERY
	SERVICE
	RPC
	RETURNS
	STREAM
	WHERE
	ORDER_BY
	LIMIT

	// SQL operators (keywords)
	AND
	OR
	NOT
	IN
	LIKE
	IS
	NULL

	// Direction
	ASC
	DESC

	// Types
	TYPE_STRING
	TYPE_INT32
	TYPE_INT64
	TYPE_FLOAT
	TYPE_DOUBLE
	TYPE_BOOL
	TYPE_BYTES
	TYPE_TIMESTAMP

	// Boolean literals
	TRUE
	FALSE
)

var tokenNames = map[TokenType]string{
	ILLEGAL:   "ILLEGAL",
	EOF:       "EOF",
	COMMENT:   "COMMENT",
	IDENT:     "IDENT",
	INT:       "INT",
	FLOAT:     "FLOAT",
	STRING:    "STRING",
	LPAREN:    "(",
	RPAREN:    ")",
	LBRACE:    "{",
	RBRACE:    "}",
	LBRACKET:  "[",
	RBRACKET:  "]",
	SEMICOLON: ";",
	COLON:     ":",
	COMMA:     ",",
	DOT:       ".",
	AT:        "@",
	QUESTION:  "?",
	EQUALS:    "=",
	BANG_EQ:   "!=",
	LT:        "<",
	LT_EQ:     "<=",
	GT:        ">",
	GT_EQ:     ">=",
	PLUS:      "+",
	MINUS:     "-",
	STAR:      "*",
	SLASH:     "/",
	PERCENT:   "%",
	CONCAT:    "||",
	PACKAGE:   "package",
	IMPORT:    "import",
	OPTION:    "option",
	ENUM:      "enum",
	ENTITY:    "entity",
	QUERY:     "query",
	SERVICE:   "service",
	RPC:       "rpc",
	RETURNS:   "returns",
	STREAM:    "stream",
	WHERE:     "where",
	ORDER_BY:  "order_by",
	LIMIT:     "limit",
	AND:       "AND",
	OR:        "OR",
	NOT:       "NOT",
	IN:        "IN",
	LIKE:      "LIKE",
	IS:        "IS",
	NULL:      "NULL",
	ASC:       "ASC",
	DESC:      "DESC",
	TYPE_STRING:    "string",
	TYPE_INT32:     "int32",
	TYPE_INT64:     "int64",
	TYPE_FLOAT:     "float",
	TYPE_DOUBLE:    "double",
	TYPE_BOOL:      "bool",
	TYPE_BYTES:     "bytes",
	TYPE_TIMESTAMP: "timestamp",
	TRUE:           "true",
	FALSE:          "false",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// keywords maps keyword strings to their token types.
var keywords = map[string]TokenType{
	"package":   PACKAGE,
	"import":    IMPORT,
	"option":    OPTION,
	"enum":      ENUM,
	"entity":    ENTITY,
	"query":     QUERY,
	"service":   SERVICE,
	"rpc":       RPC,
	"returns":   RETURNS,
	"stream":    STREAM,
	"where":     WHERE,
	"order_by":  ORDER_BY,
	"limit":     LIMIT,
	"AND":       AND,
	"OR":        OR,
	"NOT":       NOT,
	"IN":        IN,
	"LIKE":      LIKE,
	"IS":        IS,
	"NULL":      NULL,
	"ASC":       ASC,
	"DESC":      DESC,
	"string":    TYPE_STRING,
	"int32":     TYPE_INT32,
	"int64":     TYPE_INT64,
	"float":     TYPE_FLOAT,
	"double":    TYPE_DOUBLE,
	"bool":      TYPE_BOOL,
	"bytes":     TYPE_BYTES,
	"timestamp": TYPE_TIMESTAMP,
	"true":      TRUE,
	"false":     FALSE,
}

// LookupIdent returns the token type for an identifier.
// If the identifier is a keyword, returns the keyword token type.
// Otherwise, returns IDENT.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Position represents a position in the source file.
type Position struct {
	Filename string
	Line     int
	Column   int
	Offset   int
}

func (p Position) String() string {
	if p.Filename != "" {
		return p.Filename + ":" + string(rune('0'+p.Line)) + ":" + string(rune('0'+p.Column))
	}
	return string(rune('0'+p.Line)) + ":" + string(rune('0'+p.Column))
}
