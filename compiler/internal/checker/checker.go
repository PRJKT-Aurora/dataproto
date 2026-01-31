// Package checker provides semantic analysis for DataProto AST.
package checker

import (
	"fmt"
	"strings"

	"github.com/aurora/dataproto/internal/parser"
)

// Checker performs semantic analysis on a parsed DataProto file.
type Checker struct {
	file   *parser.File
	errors []Error

	// Symbol tables
	enums    map[string]*parser.EnumDecl
	entities map[string]*parser.EntityDecl
	services map[string]*parser.ServiceDecl
}

// Error represents a semantic error.
type Error struct {
	Position parser.Node
	Message  string
}

func (e Error) Error() string {
	if e.Position != nil {
		pos := e.Position.Pos()
		return fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, e.Message)
	}
	return e.Message
}

// New creates a new Checker for the given file.
func New(file *parser.File) *Checker {
	return &Checker{
		file:     file,
		enums:    make(map[string]*parser.EnumDecl),
		entities: make(map[string]*parser.EntityDecl),
		services: make(map[string]*parser.ServiceDecl),
	}
}

// Check performs semantic analysis and returns any errors.
func (c *Checker) Check() []Error {
	// Phase 1: Build symbol tables
	c.buildSymbolTables()

	// Phase 2: Check entities
	for _, entity := range c.file.Entities {
		c.checkEntity(entity)
	}

	// Phase 3: Check services
	for _, svc := range c.file.Services {
		c.checkService(svc)
	}

	return c.errors
}

func (c *Checker) addError(node parser.Node, format string, args ...interface{}) {
	c.errors = append(c.errors, Error{
		Position: node,
		Message:  fmt.Sprintf(format, args...),
	})
}

func (c *Checker) buildSymbolTables() {
	// Register enums
	for _, enum := range c.file.Enums {
		if _, exists := c.enums[enum.Name]; exists {
			c.addError(enum, "duplicate enum: %s", enum.Name)
		}
		c.enums[enum.Name] = enum
	}

	// Register entities
	for _, entity := range c.file.Entities {
		if _, exists := c.entities[entity.Name]; exists {
			c.addError(entity, "duplicate entity: %s", entity.Name)
		}
		c.entities[entity.Name] = entity
	}

	// Register services
	for _, svc := range c.file.Services {
		if _, exists := c.services[svc.Name]; exists {
			c.addError(svc, "duplicate service: %s", svc.Name)
		}
		c.services[svc.Name] = svc
	}
}

func (c *Checker) checkEntity(entity *parser.EntityDecl) {
	// Check annotations
	c.checkEntityAnnotations(entity)

	// Check fields
	fieldNames := make(map[string]bool)
	hasPrimaryKey := false

	for _, field := range entity.Fields {
		// Check duplicate field names
		if fieldNames[field.Name] {
			c.addError(field, "duplicate field: %s", field.Name)
		}
		fieldNames[field.Name] = true

		// Check field type
		c.checkType(field.Type)

		// Check field annotations
		c.checkFieldAnnotations(field)

		// Track primary key
		if field.IsPrimaryKey() {
			if hasPrimaryKey {
				c.addError(field, "entity %s has multiple primary keys", entity.Name)
			}
			hasPrimaryKey = true
		}
	}

	// Warn if no primary key
	if !hasPrimaryKey && len(entity.Fields) > 0 {
		c.addError(entity, "entity %s has no primary key (@pk)", entity.Name)
	}

	// Check queries
	for _, query := range entity.Queries {
		c.checkQuery(entity, query)
	}
}

func (c *Checker) checkEntityAnnotations(entity *parser.EntityDecl) {
	for _, ann := range entity.Annotations {
		switch ann.Name {
		case "table":
			// Check that table name is provided
			if len(ann.Args) == 0 {
				c.addError(ann, "@table requires a table name")
			} else if _, ok := ann.Args[0].Value.(string); !ok {
				c.addError(ann, "@table argument must be a string")
			}

		case "backends":
			// Check that backends are valid
			for _, arg := range ann.Args {
				if backend, ok := arg.Value.(string); ok {
					if !isValidBackend(backend) {
						c.addError(ann, "unknown backend: %s", backend)
					}
				}
			}

		default:
			c.addError(ann, "unknown entity annotation: @%s", ann.Name)
		}
	}
}

func (c *Checker) checkFieldAnnotations(field *parser.FieldDecl) {
	for _, ann := range field.Annotations {
		switch ann.Name {
		case "pk", "required", "indexed", "unique":
			// No arguments required

		case "default":
			if len(ann.Args) == 0 {
				c.addError(ann, "@default requires a value")
			}

		case "length":
			// Check for valid length arguments
			if len(ann.Args) == 0 {
				c.addError(ann, "@length requires arguments")
			}

		case "pattern":
			if len(ann.Args) == 0 {
				c.addError(ann, "@pattern requires a regex string")
			}

		case "range":
			if len(ann.Args) < 2 {
				c.addError(ann, "@range requires min and max values")
			}

		case "fk":
			if len(ann.Args) == 0 {
				c.addError(ann, "@fk requires Entity.field reference")
			} else if ref, ok := ann.Args[0].Value.(string); ok {
				parts := strings.Split(ref, ".")
				if len(parts) != 2 {
					c.addError(ann, "@fk must be in format Entity.field")
				} else if _, exists := c.entities[parts[0]]; !exists {
					c.addError(ann, "unknown entity in @fk: %s", parts[0])
				}
			}

		case "ondelete":
			if len(ann.Args) == 0 {
				c.addError(ann, "@ondelete requires action (cascade, setnull, restrict)")
			}

		default:
			c.addError(ann, "unknown field annotation: @%s", ann.Name)
		}
	}

	// Check annotation combinations
	if field.IsPrimaryKey() && field.Type.Optional {
		c.addError(field, "primary key cannot be optional")
	}
}

func (c *Checker) checkType(typeRef *parser.TypeRef) {
	// Check if type is a built-in type
	builtinTypes := map[string]bool{
		"string":    true,
		"int32":     true,
		"int64":     true,
		"float":     true,
		"double":    true,
		"bool":      true,
		"bytes":     true,
		"timestamp": true,
	}

	if builtinTypes[typeRef.Name] {
		return
	}

	// Check if type is a known enum
	if _, exists := c.enums[typeRef.Name]; exists {
		return
	}

	// Check if type is a known entity
	if _, exists := c.entities[typeRef.Name]; exists {
		return
	}

	c.addError(typeRef, "unknown type: %s", typeRef.Name)
}

func (c *Checker) checkQuery(entity *parser.EntityDecl, query *parser.QueryDecl) {
	// Build a set of valid identifiers for the query
	validIdents := make(map[string]bool)

	// Add field names
	for _, field := range entity.Fields {
		validIdents[field.Name] = true
	}

	// Add parameter names
	for _, param := range query.Params {
		validIdents[param.Name] = true
		c.checkType(param.Type)
	}

	// Check WHERE expression
	if query.Where != nil {
		c.checkExpr(query.Where, validIdents)
	}

	// Check ORDER BY fields
	for _, ob := range query.OrderBy {
		if !validIdents[ob.Field] {
			c.addError(ob, "unknown field in ORDER BY: %s", ob.Field)
		}
	}

	// Check LIMIT
	if query.Limit != nil {
		c.checkExpr(query.Limit, validIdents)
	}
}

func (c *Checker) checkExpr(expr parser.Expr, validIdents map[string]bool) {
	switch e := expr.(type) {
	case *parser.BinaryExpr:
		c.checkExpr(e.Left, validIdents)
		c.checkExpr(e.Right, validIdents)

	case *parser.UnaryExpr:
		c.checkExpr(e.Operand, validIdents)

	case *parser.IsNullExpr:
		c.checkExpr(e.Operand, validIdents)

	case *parser.IdentExpr:
		// Allow known functions and SQL keywords
		knownFunctions := map[string]bool{
			"NOW":     true,
			"COUNT":   true,
			"SUM":     true,
			"AVG":     true,
			"MIN":     true,
			"MAX":     true,
			"COALESCE": true,
		}
		if !validIdents[e.Name] && !knownFunctions[e.Name] {
			c.addError(e, "unknown identifier: %s", e.Name)
		}

	case *parser.CallExpr:
		for _, arg := range e.Args {
			c.checkExpr(arg, validIdents)
		}

	case *parser.ParenExpr:
		c.checkExpr(e.Inner, validIdents)

	case *parser.LiteralExpr:
		// Literals are always valid
	}
}

func (c *Checker) checkService(svc *parser.ServiceDecl) {
	for _, rpc := range svc.Methods {
		// Check request type
		c.checkRpcType(rpc.RequestType)

		// Check response type
		c.checkRpcType(rpc.ResponseType)
	}
}

func (c *Checker) checkRpcType(rpcType *parser.RpcType) {
	// Check if type is a known entity or a standard message type
	knownTypes := map[string]bool{
		"PushResult":    true,
		"Result":        true,
		"Empty":         true,
	}

	if _, exists := c.entities[rpcType.Name]; exists {
		return
	}

	if knownTypes[rpcType.Name] {
		return
	}

	// Check for Request/Response message types
	if strings.HasSuffix(rpcType.Name, "Request") || strings.HasSuffix(rpcType.Name, "Response") {
		return // These are typically generated
	}

	// Allow any type that starts with entity name (e.g., GetEventsRequest for CalendarEvent)
	for name := range c.entities {
		if strings.Contains(rpcType.Name, name) {
			return
		}
	}

	c.addError(rpcType, "unknown RPC type: %s", rpcType.Name)
}

func isValidBackend(backend string) bool {
	validBackends := map[string]bool{
		"sqlite":   true,
		"postgres": true,
		"ceramic":  true,
		"mysql":    true,
	}
	return validBackends[backend]
}

// Check is a convenience function to check a file.
func Check(file *parser.File) []Error {
	c := New(file)
	return c.Check()
}
