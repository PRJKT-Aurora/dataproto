package parser

import (
	"testing"
)

func TestParseCalendarSchema(t *testing.T) {
	input := `
package acos;

@table("calendar_events")
@backends(sqlite, postgres)
entity CalendarEvent {
    @pk id: string;
    @required title: string;
    @indexed start_date: timestamp;
    end_date: timestamp?;
    @default(false) is_all_day: bool;
    calendar_color: string?;
    calendar_name: string?;
    location: string?;
    notes: string?;

    query eventsByDateRange(after: timestamp, before: timestamp) {
        where start_date >= after AND start_date < before
        order_by start_date ASC
    }

    query upcomingEvents(limit: int32 = 50) {
        where start_date >= NOW()
        order_by start_date ASC
        limit limit
    }
}

service CalendarService {
    rpc PushEvents(stream CalendarEvent) returns (PushResult);
    rpc GetEvents(GetEventsRequest) returns (stream CalendarEvent);
    rpc DeleteEvent(DeleteEventRequest) returns (Result);
}
`

	file, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Check package
	if file.Package == nil {
		t.Fatal("Expected package declaration")
	}
	if file.Package.Name != "acos" {
		t.Errorf("Expected package 'acos', got '%s'", file.Package.Name)
	}

	// Check entity
	if len(file.Entities) != 1 {
		t.Fatalf("Expected 1 entity, got %d", len(file.Entities))
	}

	entity := file.Entities[0]
	if entity.Name != "CalendarEvent" {
		t.Errorf("Expected entity 'CalendarEvent', got '%s'", entity.Name)
	}

	// Check annotations
	if entity.TableName() != "calendar_events" {
		t.Errorf("Expected table name 'calendar_events', got '%s'", entity.TableName())
	}

	// Check fields
	if len(entity.Fields) != 9 {
		t.Errorf("Expected 9 fields, got %d", len(entity.Fields))
	}

	// Check first field (id)
	idField := entity.Fields[0]
	if idField.Name != "id" {
		t.Errorf("Expected first field 'id', got '%s'", idField.Name)
	}
	if !idField.IsPrimaryKey() {
		t.Error("Expected id to have @pk annotation")
	}

	// Check optional field
	endDateField := entity.Fields[3]
	if endDateField.Name != "end_date" {
		t.Errorf("Expected field 'end_date', got '%s'", endDateField.Name)
	}
	if !endDateField.Type.Optional {
		t.Error("Expected end_date to be optional")
	}

	// Check queries
	if len(entity.Queries) != 2 {
		t.Errorf("Expected 2 queries, got %d", len(entity.Queries))
	}

	query1 := entity.Queries[0]
	if query1.Name != "eventsByDateRange" {
		t.Errorf("Expected query 'eventsByDateRange', got '%s'", query1.Name)
	}
	if len(query1.Params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(query1.Params))
	}

	// Check service
	if len(file.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(file.Services))
	}

	service := file.Services[0]
	if service.Name != "CalendarService" {
		t.Errorf("Expected service 'CalendarService', got '%s'", service.Name)
	}
	if len(service.Methods) != 3 {
		t.Errorf("Expected 3 methods, got %d", len(service.Methods))
	}

	// Check streaming
	pushMethod := service.Methods[0]
	if !pushMethod.RequestType.Stream {
		t.Error("Expected PushEvents request to be streaming")
	}
}

func TestParseEnum(t *testing.T) {
	input := `
package acos;

enum MediaType {
    UNKNOWN = 0;
    IMAGE = 1;
    VIDEO = 2;
    LIVE_PHOTO = 3;
}
`

	file, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(file.Enums) != 1 {
		t.Fatalf("Expected 1 enum, got %d", len(file.Enums))
	}

	enum := file.Enums[0]
	if enum.Name != "MediaType" {
		t.Errorf("Expected enum 'MediaType', got '%s'", enum.Name)
	}

	if len(enum.Values) != 4 {
		t.Errorf("Expected 4 values, got %d", len(enum.Values))
	}

	// Check values
	expectedValues := map[string]int{
		"UNKNOWN":    0,
		"IMAGE":      1,
		"VIDEO":      2,
		"LIVE_PHOTO": 3,
	}

	for _, v := range enum.Values {
		expected, ok := expectedValues[v.Name]
		if !ok {
			t.Errorf("Unexpected enum value: %s", v.Name)
		}
		if v.Number != expected {
			t.Errorf("Expected %s = %d, got %d", v.Name, expected, v.Number)
		}
	}
}

func TestParseQueryWithComplexWhere(t *testing.T) {
	input := `
package test;

entity Item {
    @pk id: string;
    title: string;
    notes: string?;

    query search(term: string) {
        where title LIKE "%" || term || "%"
           OR notes LIKE "%" || term || "%"
        order_by title ASC
    }
}
`

	file, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	entity := file.Entities[0]
	query := entity.Queries[0]

	if query.Where == nil {
		t.Fatal("Expected WHERE clause")
	}

	// The WHERE should be an OR expression
	orExpr, ok := query.Where.(*BinaryExpr)
	if !ok {
		t.Fatalf("Expected BinaryExpr, got %T", query.Where)
	}
	if orExpr.Op != "OR" {
		t.Errorf("Expected OR, got %s", orExpr.Op)
	}
}

func TestParseAnnotationArguments(t *testing.T) {
	input := `
package test;

entity Item {
    @pk id: string;
    @length(min: 1, max: 500) title: string;
    @default(42) count: int32;
    @pattern("^[A-Z]+$") code: string;
}
`

	file, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	entity := file.Entities[0]

	// Check @length annotation
	titleField := entity.Fields[1]
	lengthAnn := titleField.GetAnnotation("length")
	if lengthAnn == nil {
		t.Fatal("Expected @length annotation on title")
	}
	if len(lengthAnn.Args) != 2 {
		t.Errorf("Expected 2 args for @length, got %d", len(lengthAnn.Args))
	}

	// Check @default annotation
	countField := entity.Fields[2]
	defaultAnn := countField.GetAnnotation("default")
	if defaultAnn == nil {
		t.Fatal("Expected @default annotation on count")
	}
	if len(defaultAnn.Args) != 1 {
		t.Errorf("Expected 1 arg for @default, got %d", len(defaultAnn.Args))
	}
	if val, ok := defaultAnn.Args[0].Value.(int64); !ok || val != 42 {
		t.Errorf("Expected default value 42, got %v", defaultAnn.Args[0].Value)
	}

	// Check @pattern annotation
	codeField := entity.Fields[3]
	patternAnn := codeField.GetAnnotation("pattern")
	if patternAnn == nil {
		t.Fatal("Expected @pattern annotation on code")
	}
	if val, ok := patternAnn.Args[0].Value.(string); !ok || val != "^[A-Z]+$" {
		t.Errorf("Expected pattern '^[A-Z]+$', got %v", patternAnn.Args[0].Value)
	}
}

func TestParseImports(t *testing.T) {
	input := `
package acos;

import "common.dataproto";
import "other/types.dataproto";

entity Test {
    @pk id: string;
}
`

	file, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(file.Imports) != 2 {
		t.Fatalf("Expected 2 imports, got %d", len(file.Imports))
	}

	if file.Imports[0].Path != "common.dataproto" {
		t.Errorf("Expected import 'common.dataproto', got '%s'", file.Imports[0].Path)
	}
	if file.Imports[1].Path != "other/types.dataproto" {
		t.Errorf("Expected import 'other/types.dataproto', got '%s'", file.Imports[1].Path)
	}
}
