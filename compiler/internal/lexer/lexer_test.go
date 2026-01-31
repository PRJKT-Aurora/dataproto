package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `package acos;

@table("calendar_events")
@backends(sqlite, postgres)
entity CalendarEvent {
    @pk id: string;
    @required title: string;
    @indexed start_date: timestamp;
    end_date: timestamp?;
    @default(false) is_all_day: bool;

    query eventsByDateRange(after: timestamp, before: timestamp) {
        where start_date >= after AND start_date < before
        order_by start_date ASC
    }
}

service CalendarService {
    rpc PushEvents(stream CalendarEvent) returns (PushResult);
}
`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{PACKAGE, "package"},
		{IDENT, "acos"},
		{SEMICOLON, ";"},
		{AT, "@"},
		{IDENT, "table"},
		{LPAREN, "("},
		{STRING, "calendar_events"},
		{RPAREN, ")"},
		{AT, "@"},
		{IDENT, "backends"},
		{LPAREN, "("},
		{IDENT, "sqlite"},
		{COMMA, ","},
		{IDENT, "postgres"},
		{RPAREN, ")"},
		{ENTITY, "entity"},
		{IDENT, "CalendarEvent"},
		{LBRACE, "{"},
		{AT, "@"},
		{IDENT, "pk"},
		{IDENT, "id"},
		{COLON, ":"},
		{TYPE_STRING, "string"},
		{SEMICOLON, ";"},
		{AT, "@"},
		{IDENT, "required"},
		{IDENT, "title"},
		{COLON, ":"},
		{TYPE_STRING, "string"},
		{SEMICOLON, ";"},
		{AT, "@"},
		{IDENT, "indexed"},
		{IDENT, "start_date"},
		{COLON, ":"},
		{TYPE_TIMESTAMP, "timestamp"},
		{SEMICOLON, ";"},
		{IDENT, "end_date"},
		{COLON, ":"},
		{TYPE_TIMESTAMP, "timestamp"},
		{QUESTION, "?"},
		{SEMICOLON, ";"},
		{AT, "@"},
		{IDENT, "default"},
		{LPAREN, "("},
		{FALSE, "false"},
		{RPAREN, ")"},
		{IDENT, "is_all_day"},
		{COLON, ":"},
		{TYPE_BOOL, "bool"},
		{SEMICOLON, ";"},
		{QUERY, "query"},
		{IDENT, "eventsByDateRange"},
		{LPAREN, "("},
		{IDENT, "after"},
		{COLON, ":"},
		{TYPE_TIMESTAMP, "timestamp"},
		{COMMA, ","},
		{IDENT, "before"},
		{COLON, ":"},
		{TYPE_TIMESTAMP, "timestamp"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{WHERE, "where"},
		{IDENT, "start_date"},
		{GT_EQ, ">="},
		{IDENT, "after"},
		{AND, "AND"},
		{IDENT, "start_date"},
		{LT, "<"},
		{IDENT, "before"},
		{ORDER_BY, "order_by"},
		{IDENT, "start_date"},
		{ASC, "ASC"},
		{RBRACE, "}"},
		{RBRACE, "}"},
		{SERVICE, "service"},
		{IDENT, "CalendarService"},
		{LBRACE, "{"},
		{RPC, "rpc"},
		{IDENT, "PushEvents"},
		{LPAREN, "("},
		{STREAM, "stream"},
		{IDENT, "CalendarEvent"},
		{RPAREN, ")"},
		{RETURNS, "returns"},
		{LPAREN, "("},
		{IDENT, "PushResult"},
		{RPAREN, ")"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNumberLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
			"123",
			[]Token{{INT, "123", 1, 1}},
		},
		{
			"-42",
			[]Token{{INT, "-42", 1, 1}},
		},
		{
			"3.14",
			[]Token{{FLOAT, "3.14", 1, 1}},
		},
		{
			"-2.5e10",
			[]Token{{FLOAT, "-2.5e10", 1, 1}},
		},
		{
			"1E-5",
			[]Token{{FLOAT, "1E-5", 1, 1}},
		},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != tt.expected[0].Type {
			t.Errorf("input=%q - type wrong. expected=%q, got=%q",
				tt.input, tt.expected[0].Type, tok.Type)
		}
		if tok.Literal != tt.expected[0].Literal {
			t.Errorf("input=%q - literal wrong. expected=%q, got=%q",
				tt.input, tt.expected[0].Literal, tok.Literal)
		}
	}
}

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`"with \"quotes\""`, `with "quotes"`},
		{`"with\nnewline"`, "with\nnewline"},
		{`"with\ttab"`, "with\ttab"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != STRING {
			t.Errorf("input=%q - expected STRING, got=%q", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input=%q - literal wrong. expected=%q, got=%q",
				tt.input, tt.expected, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `// This is a line comment
package acos;
/* This is a
   block comment */
entity Test {}`

	l := New(input)

	// Should skip comments
	tok := l.NextToken()
	if tok.Type != PACKAGE {
		t.Errorf("expected PACKAGE, got %q", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != IDENT || tok.Literal != "acos" {
		t.Errorf("expected IDENT 'acos', got %q %q", tok.Type, tok.Literal)
	}

	tok = l.NextToken() // ;
	tok = l.NextToken() // entity
	if tok.Type != ENTITY {
		t.Errorf("expected ENTITY, got %q", tok.Type)
	}
}

func TestOperators(t *testing.T) {
	input := `= != < <= > >= + - * / % ||`

	expected := []TokenType{
		EQUALS, BANG_EQ, LT, LT_EQ, GT, GT_EQ,
		PLUS, MINUS, STAR, SLASH, PERCENT, CONCAT,
	}

	l := New(input)

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Errorf("test[%d] - expected %q, got %q", i, exp, tok.Type)
		}
	}
}

func TestLineNumbers(t *testing.T) {
	input := `package
acos
;`

	l := New(input)

	tok := l.NextToken() // package
	if tok.Line != 1 {
		t.Errorf("package - expected line 1, got %d", tok.Line)
	}

	tok = l.NextToken() // acos
	if tok.Line != 2 {
		t.Errorf("acos - expected line 2, got %d", tok.Line)
	}

	tok = l.NextToken() // ;
	if tok.Line != 3 {
		t.Errorf("; - expected line 3, got %d", tok.Line)
	}
}
