package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer tokenizes DataProto source code.
type Lexer struct {
	input    string
	filename string
	pos      int  // current position in input (points to current char)
	readPos  int  // current reading position (after current char)
	ch       rune // current character under examination
	line     int  // current line number (1-indexed)
	column   int  // current column number (1-indexed)
	lineStart int // position of current line start
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	l := &Lexer{
		input:     input,
		line:      1,
		column:    1,
		lineStart: 0,
	}
	l.readChar()
	return l
}

// NewWithFilename creates a new Lexer with a filename for error messages.
func NewWithFilename(input, filename string) *Lexer {
	l := New(input)
	l.filename = filename
	return l
}

// readChar reads the next character and advances the position.
func (l *Lexer) readChar() {
	l.pos = l.readPos
	if l.readPos >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		r, width := utf8.DecodeRuneInString(l.input[l.readPos:])
		l.ch = r
		l.readPos += width
	}
	l.column = l.pos - l.lineStart + 1
}

// peekChar returns the next character without advancing.
func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return r
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespaceAndComments()

	tok := Token{
		Line:   l.line,
		Column: l.column,
	}

	switch l.ch {
	case 0:
		tok.Type = EOF
		tok.Literal = ""
	case '(':
		tok = l.newToken(LPAREN, "(")
	case ')':
		tok = l.newToken(RPAREN, ")")
	case '{':
		tok = l.newToken(LBRACE, "{")
	case '}':
		tok = l.newToken(RBRACE, "}")
	case '[':
		tok = l.newToken(LBRACKET, "[")
	case ']':
		tok = l.newToken(RBRACKET, "]")
	case ';':
		tok = l.newToken(SEMICOLON, ";")
	case ':':
		tok = l.newToken(COLON, ":")
	case ',':
		tok = l.newToken(COMMA, ",")
	case '.':
		tok = l.newToken(DOT, ".")
	case '@':
		tok = l.newToken(AT, "@")
	case '?':
		tok = l.newToken(QUESTION, "?")
	case '+':
		tok = l.newToken(PLUS, "+")
	case '-':
		if isDigit(l.peekChar()) {
			tok = l.readNumber()
		} else {
			tok = l.newToken(MINUS, "-")
		}
	case '*':
		tok = l.newToken(STAR, "*")
	case '/':
		tok = l.newToken(SLASH, "/")
	case '%':
		tok = l.newToken(PERCENT, "%")
	case '=':
		tok = l.newToken(EQUALS, "=")
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: BANG_EQ, Literal: "!=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(ILLEGAL, string(l.ch))
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: LT_EQ, Literal: "<=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(LT, "<")
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: GT_EQ, Literal: ">=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(GT, ">")
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = Token{Type: CONCAT, Literal: "||", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(ILLEGAL, string(l.ch))
		}
	case '"':
		tok = l.readString()
	default:
		if isLetter(l.ch) {
			tok = l.readIdentifier()
			return tok // return early, readIdentifier already advanced
		} else if isDigit(l.ch) {
			tok = l.readNumber()
			return tok // return early, readNumber already advanced
		} else {
			tok = l.newToken(ILLEGAL, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

// newToken creates a new token with the current position.
func (l *Lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column,
	}
}

// skipWhitespaceAndComments skips whitespace and comments.
func (l *Lexer) skipWhitespaceAndComments() {
	for {
		// Skip whitespace
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' {
			if l.ch == '\n' {
				l.line++
				l.lineStart = l.readPos
			}
			l.readChar()
		}

		// Check for comments
		if l.ch == '/' {
			if l.peekChar() == '/' {
				// Line comment
				l.skipLineComment()
				continue
			} else if l.peekChar() == '*' {
				// Block comment
				l.skipBlockComment()
				continue
			}
		}

		break
	}
}

// skipLineComment skips a // comment.
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipBlockComment skips a /* */ comment.
func (l *Lexer) skipBlockComment() {
	l.readChar() // skip '/'
	l.readChar() // skip '*'

	for {
		if l.ch == 0 {
			return // EOF
		}
		if l.ch == '\n' {
			l.line++
			l.lineStart = l.readPos
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip '*'
			l.readChar() // skip '/'
			return
		}
		l.readChar()
	}
}

// readIdentifier reads an identifier or keyword.
func (l *Lexer) readIdentifier() Token {
	startCol := l.column
	startPos := l.pos

	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	literal := l.input[startPos:l.pos]
	tokenType := LookupIdent(literal)

	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  startCol,
	}
}

// readNumber reads an integer or float literal.
func (l *Lexer) readNumber() Token {
	startCol := l.column
	startPos := l.pos
	isFloat := false

	// Handle negative sign
	if l.ch == '-' {
		l.readChar()
	}

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Check for exponent
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[startPos:l.pos]
	tokenType := INT
	if isFloat {
		tokenType = FLOAT
	}

	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  startCol,
	}
}

// readString reads a string literal.
func (l *Lexer) readString() Token {
	startCol := l.column
	var sb strings.Builder

	l.readChar() // skip opening quote

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case '"':
				sb.WriteRune('"')
			case '\\':
				sb.WriteRune('\\')
			case 'n':
				sb.WriteRune('\n')
			case 'r':
				sb.WriteRune('\r')
			case 't':
				sb.WriteRune('\t')
			case 'x':
				// Hex escape: \xHH
				l.readChar()
				hex1 := l.ch
				l.readChar()
				hex2 := l.ch
				val := hexValue(hex1)*16 + hexValue(hex2)
				sb.WriteRune(rune(val))
			default:
				sb.WriteRune(l.ch)
			}
		} else {
			if l.ch == '\n' {
				l.line++
				l.lineStart = l.readPos
			}
			sb.WriteRune(l.ch)
		}
		l.readChar()
	}

	if l.ch != '"' {
		return Token{
			Type:    ILLEGAL,
			Literal: "unterminated string",
			Line:    l.line,
			Column:  startCol,
		}
	}

	return Token{
		Type:    STRING,
		Literal: sb.String(),
		Line:    l.line,
		Column:  startCol,
	}
}

// Tokenize returns all tokens from the input.
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		tok := l.NextToken()
		if tok.Type == ILLEGAL {
			return nil, fmt.Errorf("illegal token '%s' at line %d, column %d",
				tok.Literal, tok.Line, tok.Column)
		}
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens, nil
}

// Helper functions

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func hexValue(ch rune) int {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0')
	}
	if ch >= 'a' && ch <= 'f' {
		return int(ch - 'a' + 10)
	}
	if ch >= 'A' && ch <= 'F' {
		return int(ch - 'A' + 10)
	}
	return 0
}
