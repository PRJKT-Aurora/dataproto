# DataProto

**Design the schema. Generate the plumbing.**

DataProto is a unified schema protocol that extends gRPC/protobuf axioms to include database storage semantics. Define your data model once, generate everything:

```
calendar.dataproto
       │
       ├──► .proto files ──► Any language with protoc
       ├──► SQL DDL (SQLite, Postgres)
       ├──► Java (entities, repositories, mappers, queries)
       ├──► Swift (entities, repositories, iOS mappers)
       ├──► Python (dataclasses, repositories, mappers)
       └──► Qt/C++ (QObject classes, repositories, QML-ready)
```

Stop writing boilerplate. Start designing.

## Quick Start

### 1. Define Your Schema

```dataproto
// calendar.dataproto
package acos;

@table("calendar_events")
@backends(sqlite)
entity CalendarEvent {
    @pk id: string;
    @required title: string;
    @indexed start_date: timestamp;
    end_date: timestamp?;

    query eventsByDateRange(after: timestamp, before: timestamp) {
        where start_date >= after AND start_date < before
        order_by start_date ASC
    }
}
```

### 2. Compile

```bash
dataprotoc compile \
    --input calendar.dataproto \
    --output-dir generated/ \
    --targets proto,sqlite,java
```

### 3. Use Generated Code

```java
// Generated repository with certification enforcement
CalendarEventRepository repo = new CalendarEventRepository(runtime);

// Type-safe queries
List<CalendarEvent> events = repo.eventsByDateRange(
    startTime.toEpochMilli(),
    endTime.toEpochMilli()
);
```

## Why DataProto?

### The Problem

Modern applications require maintaining three separate schema definitions:
1. **Proto files** for wire format (gRPC/protobuf)
2. **SQL DDL** for database storage
3. **Query methods** in each language

These definitions drift apart, causing bugs and maintenance burden.

### The Solution

DataProto unifies all three into a single source of truth:

```
┌──────────────────────────────────────┐
│       calendar.dataproto             │
│  (single source of truth)            │
└─────────────────┬────────────────────┘
                  │
        ┌─────────┼─────────┐
        ▼         ▼         ▼
   ┌────────┐ ┌──────┐ ┌────────┐
   │ .proto │ │ DDL  │ │ Java/  │
   │ files  │ │ SQL  │ │ Swift  │
   └────────┘ └──────┘ └────────┘
```

## Features

### Type System

| DataProto | Proto | SQLite | Java |
|-----------|-------|--------|------|
| `string` | `string` | `TEXT` | `String` |
| `int32` | `int32` | `INTEGER` | `int` |
| `int64` | `int64` | `INTEGER` | `long` |
| `timestamp` | `int64` | `INTEGER` | `long` |
| `bool` | `bool` | `INTEGER` | `boolean` |
| `T?` | `optional T` | nullable | `@Nullable T` |

### Annotations

| Annotation | Description |
|------------|-------------|
| `@table("name")` | SQL table name |
| `@pk` | Primary key |
| `@required` | NOT NULL constraint |
| `@indexed` | Create index |
| `@unique` | Unique constraint |
| `@default(value)` | Default value |
| `@length(max: n)` | String length limit |
| `@pattern("regex")` | Regex validation |
| `@range(min, max)` | Numeric range |
| `@fk(Entity.field)` | Foreign key |

### Named Queries

Define type-safe queries in the schema:

```dataproto
entity Reminder {
    @pk id: string;
    @required title: string;
    due_date: timestamp?;
    @default(false) is_completed: bool;

    query incomplete(limit: int32 = 100) {
        where is_completed = false
        order_by due_date ASC
        limit limit
    }
}
```

Generated Java:

```java
public List<Reminder> incomplete(Integer limit) {
    // Type-safe SQL execution
}
```

## Certification

DataProto requires certification to use in production. This ensures all implementations uphold user-centric data principles:

1. **Data Sovereignty** - User data in user-controlled location
2. **No Surveillance** - No behavioral tracking
3. **Export/Delete** - Users can export and delete data
4. **Interoperability** - Works with standard protocols
5. **No Extraction** - Data stays on user's device

See [certification/principles.md](certification/principles.md) for details.

### Development Mode

For local development:

```java
DataProtoRuntime runtime = DataProtoRuntime.development("test.db");
```

### Production

```java
Certificate cert = Certificate.load("dataproto.cert");
DataProtoRuntime runtime = DataProtoRuntime.builder()
    .databasePath("app_data.db")
    .certificate(cert)
    .build();
```

## Project Structure

```
dataproto/
├── spec/                  # Language specification
│   └── grammar.ebnf       # Formal grammar
├── compiler/              # Go compiler (dataprotoc)
│   ├── cmd/dataprotoc/    # CLI entry point
│   └── internal/          # Lexer, parser, codegen
├── runtime/               # Language runtimes
│   ├── java/              # Java runtime with certification
│   ├── swift/             # Swift runtime (planned)
│   └── python/            # Python runtime (planned)
├── certification/         # Certification infrastructure
│   ├── principles.md      # What implementations agree to
│   └── test_suite/        # Certification tests
└── examples/aurora/       # Aurora schemas as reference
```

## Building

```bash
cd compiler
go build -o bin/dataprotoc ./cmd/dataprotoc
```

## Status

- [x] Language specification (grammar.ebnf)
- [x] Lexer + Parser + Type Checker
- [x] Proto generator
- [x] SQLite DDL generator
- [x] Java codegen (entity, repository, mapper, queries)
- [x] Swift codegen (entity, repository, iOS mappers for EKEvent/EKReminder/PHAsset)
- [x] Python codegen (dataclass, repository, mapper)
- [x] Qt/C++ codegen (QObject, repository, QML-ready, CMakeLists.txt)
- [x] Java runtime with certification
- [ ] Postgres DDL generator
- [ ] Swift runtime with certification
- [ ] Python runtime with certification
- [ ] Qt runtime with certification
- [ ] Migration differ
- [ ] VS Code extension
- [ ] Certification test suite

## License

DataProto is source-available with certification requirements. See LICENSE.md for details.

Commercial use requires certification. Apply at https://dataproto.dev/certify
