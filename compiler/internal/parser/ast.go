// Package parser provides parsing for DataProto schema files.
package parser

import "github.com/aurora/dataproto/internal/lexer"

// Node is the base interface for all AST nodes.
type Node interface {
	node()
	Pos() lexer.Position
}

// File represents a complete DataProto schema file.
type File struct {
	Position   lexer.Position
	Package    *PackageDecl
	Imports    []*ImportDecl
	Options    []*OptionDecl
	Enums      []*EnumDecl
	Entities   []*EntityDecl
	Services   []*ServiceDecl
}

func (f *File) node() {}
func (f *File) Pos() lexer.Position { return f.Position }

// PackageDecl represents a package declaration.
type PackageDecl struct {
	Position lexer.Position
	Name     string // e.g., "acos" or "acos.calendar"
}

func (p *PackageDecl) node() {}
func (p *PackageDecl) Pos() lexer.Position { return p.Position }

// ImportDecl represents an import declaration.
type ImportDecl struct {
	Position lexer.Position
	Path     string // e.g., "common.dataproto"
}

func (i *ImportDecl) node() {}
func (i *ImportDecl) Pos() lexer.Position { return i.Position }

// OptionDecl represents a file-level option.
type OptionDecl struct {
	Position lexer.Position
	Name     string
	Value    interface{} // string, int, float, bool, or identifier
}

func (o *OptionDecl) node() {}
func (o *OptionDecl) Pos() lexer.Position { return o.Position }

// EnumDecl represents an enum declaration.
type EnumDecl struct {
	Position lexer.Position
	Name     string
	Values   []*EnumValue
}

func (e *EnumDecl) node() {}
func (e *EnumDecl) Pos() lexer.Position { return e.Position }

// EnumValue represents a single enum value.
type EnumValue struct {
	Position lexer.Position
	Name     string
	Number   int
}

func (e *EnumValue) node() {}
func (e *EnumValue) Pos() lexer.Position { return e.Position }

// EntityDecl represents an entity declaration (maps to table + proto message).
type EntityDecl struct {
	Position    lexer.Position
	Annotations []*Annotation
	Name        string
	Fields      []*FieldDecl
	Queries     []*QueryDecl
}

func (e *EntityDecl) node() {}
func (e *EntityDecl) Pos() lexer.Position { return e.Position }

// Annotation represents an annotation like @table("name").
type Annotation struct {
	Position lexer.Position
	Name     string
	Args     []AnnotationArg
}

func (a *Annotation) node() {}
func (a *Annotation) Pos() lexer.Position { return a.Position }

// AnnotationArg represents an argument to an annotation.
type AnnotationArg struct {
	Position lexer.Position
	Name     string      // optional, for named args like max: 100
	Value    interface{} // string, int, float, bool, identifier, or []interface{}
}

func (a *AnnotationArg) node() {}
func (a *AnnotationArg) Pos() lexer.Position { return a.Position }

// FieldDecl represents a field in an entity.
type FieldDecl struct {
	Position    lexer.Position
	Annotations []*Annotation
	Name        string
	Type        *TypeRef
}

func (f *FieldDecl) node() {}
func (f *FieldDecl) Pos() lexer.Position { return f.Position }

// TypeRef represents a type reference.
type TypeRef struct {
	Position lexer.Position
	Name     string // base type name (string, int32, etc. or custom type)
	Optional bool   // true if followed by ?
}

func (t *TypeRef) node() {}
func (t *TypeRef) Pos() lexer.Position { return t.Position }

// QueryDecl represents a named query within an entity.
type QueryDecl struct {
	Position lexer.Position
	Name     string
	Params   []*QueryParam
	Where    Expr
	OrderBy  []*OrderByField
	Limit    Expr // can be nil, int literal, or parameter reference
}

func (q *QueryDecl) node() {}
func (q *QueryDecl) Pos() lexer.Position { return q.Position }

// QueryParam represents a parameter to a query.
type QueryParam struct {
	Position lexer.Position
	Name     string
	Type     *TypeRef
	Default  interface{} // optional default value
}

func (q *QueryParam) node() {}
func (q *QueryParam) Pos() lexer.Position { return q.Position }

// OrderByField represents a field in ORDER BY clause.
type OrderByField struct {
	Position   lexer.Position
	Field      string
	Descending bool
}

func (o *OrderByField) node() {}
func (o *OrderByField) Pos() lexer.Position { return o.Position }

// Expr is the interface for all expression types.
type Expr interface {
	Node
	expr()
}

// BinaryExpr represents a binary expression (e.g., a AND b, x >= y).
type BinaryExpr struct {
	Position lexer.Position
	Left     Expr
	Op       string // AND, OR, =, !=, <, <=, >, >=, LIKE, IN, +, -, *, /, %, ||
	Right    Expr
}

func (b *BinaryExpr) node() {}
func (b *BinaryExpr) expr() {}
func (b *BinaryExpr) Pos() lexer.Position { return b.Position }

// UnaryExpr represents a unary expression (e.g., NOT x, -5).
type UnaryExpr struct {
	Position lexer.Position
	Op       string // NOT, -
	Operand  Expr
}

func (u *UnaryExpr) node() {}
func (u *UnaryExpr) expr() {}
func (u *UnaryExpr) Pos() lexer.Position { return u.Position }

// IsNullExpr represents an IS NULL or IS NOT NULL expression.
type IsNullExpr struct {
	Position lexer.Position
	Operand  Expr
	Not      bool // true for IS NOT NULL
}

func (i *IsNullExpr) node() {}
func (i *IsNullExpr) expr() {}
func (i *IsNullExpr) Pos() lexer.Position { return i.Position }

// IdentExpr represents an identifier reference.
type IdentExpr struct {
	Position lexer.Position
	Name     string
}

func (i *IdentExpr) node() {}
func (i *IdentExpr) expr() {}
func (i *IdentExpr) Pos() lexer.Position { return i.Position }

// LiteralExpr represents a literal value.
type LiteralExpr struct {
	Position lexer.Position
	Value    interface{} // string, int64, float64, bool
}

func (l *LiteralExpr) node() {}
func (l *LiteralExpr) expr() {}
func (l *LiteralExpr) Pos() lexer.Position { return l.Position }

// CallExpr represents a function call.
type CallExpr struct {
	Position lexer.Position
	Name     string
	Args     []Expr
}

func (c *CallExpr) node() {}
func (c *CallExpr) expr() {}
func (c *CallExpr) Pos() lexer.Position { return c.Position }

// ParenExpr represents a parenthesized expression.
type ParenExpr struct {
	Position lexer.Position
	Inner    Expr
}

func (p *ParenExpr) node() {}
func (p *ParenExpr) expr() {}
func (p *ParenExpr) Pos() lexer.Position { return p.Position }

// ServiceDecl represents a gRPC service declaration.
type ServiceDecl struct {
	Position lexer.Position
	Name     string
	Methods  []*RpcDecl
}

func (s *ServiceDecl) node() {}
func (s *ServiceDecl) Pos() lexer.Position { return s.Position }

// RpcDecl represents an RPC method declaration.
type RpcDecl struct {
	Position       lexer.Position
	Name           string
	RequestType    *RpcType
	ResponseType   *RpcType
}

func (r *RpcDecl) node() {}
func (r *RpcDecl) Pos() lexer.Position { return r.Position }

// RpcType represents a request or response type in an RPC.
type RpcType struct {
	Position lexer.Position
	Stream   bool
	Name     string
}

func (r *RpcType) node() {}
func (r *RpcType) Pos() lexer.Position { return r.Position }

// Helper methods for common operations

// GetAnnotation returns the first annotation with the given name, or nil.
func (e *EntityDecl) GetAnnotation(name string) *Annotation {
	for _, a := range e.Annotations {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// GetAnnotation returns the first annotation with the given name, or nil.
func (f *FieldDecl) GetAnnotation(name string) *Annotation {
	for _, a := range f.Annotations {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// HasAnnotation returns true if the field has the given annotation.
func (f *FieldDecl) HasAnnotation(name string) bool {
	return f.GetAnnotation(name) != nil
}

// IsPrimaryKey returns true if the field has the @pk annotation.
func (f *FieldDecl) IsPrimaryKey() bool {
	return f.HasAnnotation("pk")
}

// IsRequired returns true if the field has the @required annotation.
func (f *FieldDecl) IsRequired() bool {
	return f.HasAnnotation("required")
}

// IsIndexed returns true if the field has the @indexed annotation.
func (f *FieldDecl) IsIndexed() bool {
	return f.HasAnnotation("indexed")
}

// IsUnique returns true if the field has the @unique annotation.
func (f *FieldDecl) IsUnique() bool {
	return f.HasAnnotation("unique")
}

// TableName returns the SQL table name from @table annotation, or empty string.
func (e *EntityDecl) TableName() string {
	if a := e.GetAnnotation("table"); a != nil && len(a.Args) > 0 {
		if s, ok := a.Args[0].Value.(string); ok {
			return s
		}
	}
	return ""
}

// Backends returns the list of backends from @backends annotation.
func (e *EntityDecl) Backends() []string {
	if a := e.GetAnnotation("backends"); a != nil {
		var backends []string
		for _, arg := range a.Args {
			if s, ok := arg.Value.(string); ok {
				backends = append(backends, s)
			}
		}
		return backends
	}
	return nil
}
