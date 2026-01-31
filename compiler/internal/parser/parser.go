package parser

import (
	"fmt"
	"strconv"

	"github.com/aurora/dataproto/internal/lexer"
)

// Parser parses DataProto source code into an AST.
type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []string
	filename  string
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// Read two tokens to populate curToken and peekToken
	p.nextToken()
	p.nextToken()
	return p
}

// NewFromString creates a new Parser for the given input string.
func NewFromString(input string) *Parser {
	return New(lexer.New(input))
}

// NewFromStringWithFilename creates a new Parser with a filename for error messages.
func NewFromStringWithFilename(input, filename string) *Parser {
	p := New(lexer.NewWithFilename(input, filename))
	p.filename = filename
	return p
}

// Errors returns all parsing errors.
func (p *Parser) Errors() []string {
	return p.errors
}

// nextToken advances to the next token.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs returns true if the current token is of the given type.
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs returns true if the peek token is of the given type.
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek advances if the peek token matches, otherwise adds an error.
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekError adds an error for unexpected peek token.
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("line %d:%d: expected %s, got %s",
		p.peekToken.Line, p.peekToken.Column, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// curError adds an error for unexpected current token.
func (p *Parser) curError(expected string) {
	msg := fmt.Sprintf("line %d:%d: expected %s, got %s",
		p.curToken.Line, p.curToken.Column, expected, p.curToken.Type)
	p.errors = append(p.errors, msg)
}

// curPos returns the current token position.
func (p *Parser) curPos() lexer.Position {
	return lexer.Position{
		Filename: p.filename,
		Line:     p.curToken.Line,
		Column:   p.curToken.Column,
	}
}

// isKeywordAsIdent returns true if current token is a keyword that can be used as identifier.
func (p *Parser) isKeywordAsIdent() bool {
	switch p.curToken.Type {
	case lexer.LIMIT, lexer.WHERE, lexer.ORDER_BY, lexer.QUERY,
		lexer.ASC, lexer.DESC, lexer.AND, lexer.OR, lexer.NOT,
		lexer.IN, lexer.LIKE, lexer.IS, lexer.NULL:
		return true
	default:
		return false
	}
}

// ParseFile parses a complete DataProto file.
func (p *Parser) ParseFile() *File {
	file := &File{Position: p.curPos()}

	for !p.curTokenIs(lexer.EOF) {
		switch p.curToken.Type {
		case lexer.PACKAGE:
			file.Package = p.parsePackageDecl()
		case lexer.IMPORT:
			file.Imports = append(file.Imports, p.parseImportDecl())
		case lexer.OPTION:
			file.Options = append(file.Options, p.parseOptionDecl())
		case lexer.ENUM:
			file.Enums = append(file.Enums, p.parseEnumDecl())
		case lexer.AT:
			// Annotation followed by entity or other declaration
			annotations := p.parseAnnotations()
			if p.curTokenIs(lexer.ENTITY) {
				entity := p.parseEntityDecl()
				entity.Annotations = annotations
				file.Entities = append(file.Entities, entity)
			} else {
				p.curError("entity after annotations")
				p.nextToken()
			}
		case lexer.ENTITY:
			file.Entities = append(file.Entities, p.parseEntityDecl())
		case lexer.SERVICE:
			file.Services = append(file.Services, p.parseServiceDecl())
		default:
			p.curError("package, import, option, enum, entity, or service")
			p.nextToken()
		}
	}

	return file
}

// parsePackageDecl parses: package name.space;
func (p *Parser) parsePackageDecl() *PackageDecl {
	decl := &PackageDecl{Position: p.curPos()}
	p.nextToken() // consume 'package'

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("package name")
		return decl
	}

	decl.Name = p.curToken.Literal
	p.nextToken()

	// Handle dotted names: acos.calendar
	for p.curTokenIs(lexer.DOT) {
		p.nextToken() // consume '.'
		if !p.curTokenIs(lexer.IDENT) {
			p.curError("identifier after '.'")
			break
		}
		decl.Name += "." + p.curToken.Literal
		p.nextToken()
	}

	if p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken() // consume ';'
	}

	return decl
}

// parseImportDecl parses: import "path";
func (p *Parser) parseImportDecl() *ImportDecl {
	decl := &ImportDecl{Position: p.curPos()}
	p.nextToken() // consume 'import'

	if !p.curTokenIs(lexer.STRING) {
		p.curError("import path string")
		return decl
	}

	decl.Path = p.curToken.Literal
	p.nextToken()

	if p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return decl
}

// parseOptionDecl parses: option name = value;
func (p *Parser) parseOptionDecl() *OptionDecl {
	decl := &OptionDecl{Position: p.curPos()}
	p.nextToken() // consume 'option'

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("option name")
		return decl
	}

	decl.Name = p.curToken.Literal
	p.nextToken()

	if !p.curTokenIs(lexer.EQUALS) {
		p.curError("'='")
		return decl
	}
	p.nextToken()

	decl.Value = p.parseValue()
	p.nextToken()

	if p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return decl
}

// parseEnumDecl parses: enum Name { VALUE = 0; ... }
func (p *Parser) parseEnumDecl() *EnumDecl {
	decl := &EnumDecl{Position: p.curPos()}
	p.nextToken() // consume 'enum'

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("enum name")
		return decl
	}

	decl.Name = p.curToken.Literal
	p.nextToken()

	if !p.curTokenIs(lexer.LBRACE) {
		p.curError("'{'")
		return decl
	}
	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.IDENT) {
			value := &EnumValue{Position: p.curPos(), Name: p.curToken.Literal}
			p.nextToken()

			if p.curTokenIs(lexer.EQUALS) {
				p.nextToken()
				if p.curTokenIs(lexer.INT) {
					num, _ := strconv.Atoi(p.curToken.Literal)
					value.Number = num
					p.nextToken()
				}
			}

			if p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
			}

			decl.Values = append(decl.Values, value)
		} else {
			p.curError("enum value name")
			p.nextToken()
		}
	}

	if p.curTokenIs(lexer.RBRACE) {
		p.nextToken()
	}

	return decl
}

// parseEntityDecl parses: entity Name { fields... queries... }
func (p *Parser) parseEntityDecl() *EntityDecl {
	decl := &EntityDecl{Position: p.curPos()}
	p.nextToken() // consume 'entity'

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("entity name")
		return decl
	}

	decl.Name = p.curToken.Literal
	p.nextToken()

	if !p.curTokenIs(lexer.LBRACE) {
		p.curError("'{'")
		return decl
	}
	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		switch {
		case p.curTokenIs(lexer.AT):
			// Annotated field
			annotations := p.parseAnnotations()
			if p.curTokenIs(lexer.IDENT) {
				field := p.parseFieldDecl()
				field.Annotations = annotations
				decl.Fields = append(decl.Fields, field)
			}
		case p.curTokenIs(lexer.IDENT):
			decl.Fields = append(decl.Fields, p.parseFieldDecl())
		case p.curTokenIs(lexer.QUERY):
			decl.Queries = append(decl.Queries, p.parseQueryDecl())
		default:
			p.curError("field, query, or '}'")
			p.nextToken()
		}
	}

	if p.curTokenIs(lexer.RBRACE) {
		p.nextToken()
	}

	return decl
}

// parseAnnotations parses a sequence of @annotation(args).
func (p *Parser) parseAnnotations() []*Annotation {
	var annotations []*Annotation

	for p.curTokenIs(lexer.AT) {
		annotations = append(annotations, p.parseAnnotation())
	}

	return annotations
}

// parseAnnotation parses: @name or @name(args)
func (p *Parser) parseAnnotation() *Annotation {
	ann := &Annotation{Position: p.curPos()}
	p.nextToken() // consume '@'

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("annotation name")
		return ann
	}

	ann.Name = p.curToken.Literal
	p.nextToken()

	// Optional arguments
	if p.curTokenIs(lexer.LPAREN) {
		p.nextToken()

		for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
			arg := p.parseAnnotationArg()
			ann.Args = append(ann.Args, arg)

			if p.curTokenIs(lexer.COMMA) {
				p.nextToken()
			} else {
				break
			}
		}

		if p.curTokenIs(lexer.RPAREN) {
			p.nextToken()
		}
	}

	return ann
}

// parseAnnotationArg parses an annotation argument.
func (p *Parser) parseAnnotationArg() AnnotationArg {
	arg := AnnotationArg{Position: p.curPos()}

	// Check for named argument: name = value or name: value
	if p.curTokenIs(lexer.IDENT) && (p.peekTokenIs(lexer.EQUALS) || p.peekTokenIs(lexer.COLON)) {
		arg.Name = p.curToken.Literal
		p.nextToken() // consume name
		p.nextToken() // consume = or :
	}

	arg.Value = p.parseAnnotationValue()
	return arg
}

// parseAnnotationValue parses an annotation value.
func (p *Parser) parseAnnotationValue() interface{} {
	switch p.curToken.Type {
	case lexer.STRING:
		val := p.curToken.Literal
		p.nextToken()
		return val
	case lexer.INT:
		val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
		p.nextToken()
		return val
	case lexer.FLOAT:
		val, _ := strconv.ParseFloat(p.curToken.Literal, 64)
		p.nextToken()
		return val
	case lexer.TRUE:
		p.nextToken()
		return true
	case lexer.FALSE:
		p.nextToken()
		return false
	case lexer.IDENT:
		val := p.curToken.Literal
		p.nextToken()
		return val
	case lexer.LBRACKET:
		return p.parseAnnotationList()
	default:
		p.nextToken()
		return nil
	}
}

// parseAnnotationList parses: [value, value, ...]
func (p *Parser) parseAnnotationList() []interface{} {
	p.nextToken() // consume '['
	var values []interface{}

	for !p.curTokenIs(lexer.RBRACKET) && !p.curTokenIs(lexer.EOF) {
		values = append(values, p.parseAnnotationValue())
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
		}
	}

	if p.curTokenIs(lexer.RBRACKET) {
		p.nextToken()
	}

	return values
}

// parseFieldDecl parses: name: Type;
func (p *Parser) parseFieldDecl() *FieldDecl {
	field := &FieldDecl{Position: p.curPos()}

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("field name")
		return field
	}

	field.Name = p.curToken.Literal
	p.nextToken()

	if !p.curTokenIs(lexer.COLON) {
		p.curError("':'")
		return field
	}
	p.nextToken()

	field.Type = p.parseTypeRef()

	if p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return field
}

// parseTypeRef parses a type reference like string, int32?, etc.
func (p *Parser) parseTypeRef() *TypeRef {
	typeRef := &TypeRef{Position: p.curPos()}

	// Check for built-in types
	switch p.curToken.Type {
	case lexer.TYPE_STRING:
		typeRef.Name = "string"
	case lexer.TYPE_INT32:
		typeRef.Name = "int32"
	case lexer.TYPE_INT64:
		typeRef.Name = "int64"
	case lexer.TYPE_FLOAT:
		typeRef.Name = "float"
	case lexer.TYPE_DOUBLE:
		typeRef.Name = "double"
	case lexer.TYPE_BOOL:
		typeRef.Name = "bool"
	case lexer.TYPE_BYTES:
		typeRef.Name = "bytes"
	case lexer.TYPE_TIMESTAMP:
		typeRef.Name = "timestamp"
	case lexer.IDENT:
		typeRef.Name = p.curToken.Literal
	default:
		p.curError("type name")
		return typeRef
	}

	p.nextToken()

	// Check for optional marker
	if p.curTokenIs(lexer.QUESTION) {
		typeRef.Optional = true
		p.nextToken()
	}

	return typeRef
}

// parseQueryDecl parses: query name(params) { where... order_by... limit... }
func (p *Parser) parseQueryDecl() *QueryDecl {
	query := &QueryDecl{Position: p.curPos()}
	p.nextToken() // consume 'query'

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("query name")
		return query
	}

	query.Name = p.curToken.Literal
	p.nextToken()

	// Parse parameters
	if !p.curTokenIs(lexer.LPAREN) {
		p.curError("'('")
		return query
	}
	p.nextToken()

	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		param := p.parseQueryParam()
		query.Params = append(query.Params, param)

		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	if p.curTokenIs(lexer.RPAREN) {
		p.nextToken()
	}

	// Parse body
	if !p.curTokenIs(lexer.LBRACE) {
		p.curError("'{'")
		return query
	}
	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		switch p.curToken.Type {
		case lexer.WHERE:
			p.nextToken()
			query.Where = p.parseExpression()
		case lexer.ORDER_BY:
			p.nextToken()
			query.OrderBy = p.parseOrderBy()
		case lexer.LIMIT:
			p.nextToken()
			query.Limit = p.parsePrimaryExpr()
		default:
			p.curError("where, order_by, limit, or '}'")
			p.nextToken()
		}
	}

	if p.curTokenIs(lexer.RBRACE) {
		p.nextToken()
	}

	return query
}

// parseQueryParam parses: name: Type = default
func (p *Parser) parseQueryParam() *QueryParam {
	param := &QueryParam{Position: p.curPos()}

	// Allow keywords to be used as parameter names (e.g., "limit")
	if !p.curTokenIs(lexer.IDENT) && !p.isKeywordAsIdent() {
		p.curError("parameter name")
		return param
	}

	param.Name = p.curToken.Literal
	p.nextToken()

	if !p.curTokenIs(lexer.COLON) {
		p.curError("':'")
		return param
	}
	p.nextToken()

	param.Type = p.parseTypeRef()

	// Optional default value
	if p.curTokenIs(lexer.EQUALS) {
		p.nextToken()
		param.Default = p.parseValue()
		p.nextToken()
	}

	return param
}

// parseOrderBy parses: field ASC, field2 DESC
func (p *Parser) parseOrderBy() []*OrderByField {
	var fields []*OrderByField

	for {
		field := &OrderByField{Position: p.curPos()}

		if !p.curTokenIs(lexer.IDENT) {
			break
		}

		field.Field = p.curToken.Literal
		p.nextToken()

		if p.curTokenIs(lexer.ASC) {
			field.Descending = false
			p.nextToken()
		} else if p.curTokenIs(lexer.DESC) {
			field.Descending = true
			p.nextToken()
		}

		fields = append(fields, field)

		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	return fields
}

// parseExpression parses a full expression (OR has lowest precedence).
func (p *Parser) parseExpression() Expr {
	return p.parseOrExpr()
}

// parseOrExpr parses: expr OR expr
func (p *Parser) parseOrExpr() Expr {
	left := p.parseAndExpr()

	for p.curTokenIs(lexer.OR) {
		op := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()
		right := p.parseAndExpr()
		left = &BinaryExpr{Position: pos, Left: left, Op: op, Right: right}
	}

	return left
}

// parseAndExpr parses: expr AND expr
func (p *Parser) parseAndExpr() Expr {
	left := p.parseCompareExpr()

	for p.curTokenIs(lexer.AND) {
		op := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()
		right := p.parseCompareExpr()
		left = &BinaryExpr{Position: pos, Left: left, Op: op, Right: right}
	}

	return left
}

// parseCompareExpr parses comparison expressions.
func (p *Parser) parseCompareExpr() Expr {
	left := p.parseAddExpr()

	switch p.curToken.Type {
	case lexer.EQUALS, lexer.BANG_EQ, lexer.LT, lexer.LT_EQ, lexer.GT, lexer.GT_EQ:
		op := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()
		right := p.parseAddExpr()
		return &BinaryExpr{Position: pos, Left: left, Op: op, Right: right}

	case lexer.LIKE:
		op := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()
		right := p.parseAddExpr()
		return &BinaryExpr{Position: pos, Left: left, Op: op, Right: right}

	case lexer.IN:
		op := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()
		right := p.parseAddExpr()
		return &BinaryExpr{Position: pos, Left: left, Op: op, Right: right}

	case lexer.IS:
		pos := p.curPos()
		p.nextToken()
		notNull := false
		if p.curTokenIs(lexer.NOT) {
			notNull = true
			p.nextToken()
		}
		if p.curTokenIs(lexer.NULL) {
			p.nextToken()
		}
		return &IsNullExpr{Position: pos, Operand: left, Not: notNull}
	}

	return left
}

// parseAddExpr parses addition/subtraction/concatenation.
func (p *Parser) parseAddExpr() Expr {
	left := p.parseMulExpr()

	for p.curTokenIs(lexer.PLUS) || p.curTokenIs(lexer.MINUS) || p.curTokenIs(lexer.CONCAT) {
		op := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()
		right := p.parseMulExpr()
		left = &BinaryExpr{Position: pos, Left: left, Op: op, Right: right}
	}

	return left
}

// parseMulExpr parses multiplication/division/modulo.
func (p *Parser) parseMulExpr() Expr {
	left := p.parseUnaryExpr()

	for p.curTokenIs(lexer.STAR) || p.curTokenIs(lexer.SLASH) || p.curTokenIs(lexer.PERCENT) {
		op := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()
		right := p.parseUnaryExpr()
		left = &BinaryExpr{Position: pos, Left: left, Op: op, Right: right}
	}

	return left
}

// parseUnaryExpr parses: NOT expr or -expr
func (p *Parser) parseUnaryExpr() Expr {
	if p.curTokenIs(lexer.NOT) {
		pos := p.curPos()
		p.nextToken()
		operand := p.parseUnaryExpr()
		return &UnaryExpr{Position: pos, Op: "NOT", Operand: operand}
	}

	if p.curTokenIs(lexer.MINUS) {
		pos := p.curPos()
		p.nextToken()
		operand := p.parseUnaryExpr()
		return &UnaryExpr{Position: pos, Op: "-", Operand: operand}
	}

	return p.parsePrimaryExpr()
}

// parsePrimaryExpr parses primary expressions.
func (p *Parser) parsePrimaryExpr() Expr {
	// Handle keywords that can be used as identifiers in expressions
	if p.isKeywordAsIdent() {
		name := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()
		return &IdentExpr{Position: pos, Name: name}
	}

	switch p.curToken.Type {
	case lexer.IDENT:
		name := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()

		// Check for function call
		if p.curTokenIs(lexer.LPAREN) {
			return p.parseCallExpr(name, pos)
		}

		return &IdentExpr{Position: pos, Name: name}

	case lexer.INT:
		val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
		pos := p.curPos()
		p.nextToken()
		return &LiteralExpr{Position: pos, Value: val}

	case lexer.FLOAT:
		val, _ := strconv.ParseFloat(p.curToken.Literal, 64)
		pos := p.curPos()
		p.nextToken()
		return &LiteralExpr{Position: pos, Value: val}

	case lexer.STRING:
		val := p.curToken.Literal
		pos := p.curPos()
		p.nextToken()
		return &LiteralExpr{Position: pos, Value: val}

	case lexer.TRUE:
		pos := p.curPos()
		p.nextToken()
		return &LiteralExpr{Position: pos, Value: true}

	case lexer.FALSE:
		pos := p.curPos()
		p.nextToken()
		return &LiteralExpr{Position: pos, Value: false}

	case lexer.LPAREN:
		pos := p.curPos()
		p.nextToken()
		inner := p.parseExpression()
		if p.curTokenIs(lexer.RPAREN) {
			p.nextToken()
		}
		return &ParenExpr{Position: pos, Inner: inner}

	default:
		pos := p.curPos()
		p.curError("expression")
		p.nextToken()
		return &LiteralExpr{Position: pos, Value: nil}
	}
}

// parseCallExpr parses: name(arg, arg, ...)
func (p *Parser) parseCallExpr(name string, pos lexer.Position) Expr {
	call := &CallExpr{Position: pos, Name: name}
	p.nextToken() // consume '('

	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		arg := p.parseExpression()
		call.Args = append(call.Args, arg)

		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	if p.curTokenIs(lexer.RPAREN) {
		p.nextToken()
	}

	return call
}

// parseServiceDecl parses: service Name { rpc methods... }
func (p *Parser) parseServiceDecl() *ServiceDecl {
	svc := &ServiceDecl{Position: p.curPos()}
	p.nextToken() // consume 'service'

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("service name")
		return svc
	}

	svc.Name = p.curToken.Literal
	p.nextToken()

	if !p.curTokenIs(lexer.LBRACE) {
		p.curError("'{'")
		return svc
	}
	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.RPC) {
			svc.Methods = append(svc.Methods, p.parseRpcDecl())
		} else {
			p.curError("rpc or '}'")
			p.nextToken()
		}
	}

	if p.curTokenIs(lexer.RBRACE) {
		p.nextToken()
	}

	return svc
}

// parseRpcDecl parses: rpc Name(Type) returns (Type);
func (p *Parser) parseRpcDecl() *RpcDecl {
	rpc := &RpcDecl{Position: p.curPos()}
	p.nextToken() // consume 'rpc'

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("rpc name")
		return rpc
	}

	rpc.Name = p.curToken.Literal
	p.nextToken()

	// Request type
	if !p.curTokenIs(lexer.LPAREN) {
		p.curError("'('")
		return rpc
	}
	p.nextToken()

	rpc.RequestType = p.parseRpcType()

	if !p.curTokenIs(lexer.RPAREN) {
		p.curError("')'")
		return rpc
	}
	p.nextToken()

	// returns
	if !p.curTokenIs(lexer.RETURNS) {
		p.curError("'returns'")
		return rpc
	}
	p.nextToken()

	// Response type
	if !p.curTokenIs(lexer.LPAREN) {
		p.curError("'('")
		return rpc
	}
	p.nextToken()

	rpc.ResponseType = p.parseRpcType()

	if !p.curTokenIs(lexer.RPAREN) {
		p.curError("')'")
		return rpc
	}
	p.nextToken()

	if p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return rpc
}

// parseRpcType parses: [stream] TypeName
func (p *Parser) parseRpcType() *RpcType {
	rpcType := &RpcType{Position: p.curPos()}

	if p.curTokenIs(lexer.STREAM) {
		rpcType.Stream = true
		p.nextToken()
	}

	if !p.curTokenIs(lexer.IDENT) {
		p.curError("type name")
		return rpcType
	}

	rpcType.Name = p.curToken.Literal
	p.nextToken()

	return rpcType
}

// parseValue parses a literal value.
func (p *Parser) parseValue() interface{} {
	switch p.curToken.Type {
	case lexer.STRING:
		return p.curToken.Literal
	case lexer.INT:
		val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
		return val
	case lexer.FLOAT:
		val, _ := strconv.ParseFloat(p.curToken.Literal, 64)
		return val
	case lexer.TRUE:
		return true
	case lexer.FALSE:
		return false
	case lexer.IDENT:
		return p.curToken.Literal
	default:
		return nil
	}
}

// Parse is a convenience function to parse a string.
func Parse(input string) (*File, error) {
	p := NewFromString(input)
	file := p.ParseFile()
	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.errors)
	}
	return file, nil
}

// ParseFile is a convenience function to parse a file.
func ParseFile(input, filename string) (*File, error) {
	p := NewFromStringWithFilename(input, filename)
	file := p.ParseFile()
	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.errors)
	}
	return file, nil
}
