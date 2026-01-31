# DataProto Implementation Guide

## Overview

DataProto is a unified schema protocol that extends gRPC/protobuf axioms to include database storage semantics. This document tracks the implementation progress and provides detailed specifications for each component.

---

## Implementation Status

### Phase 1: Language Design + Parser
| Task | Status | File |
|------|--------|------|
| Define formal grammar (EBNF) | ✅ DONE | `spec/grammar.ebnf` |
| Implement Go lexer | ✅ DONE | `compiler/internal/lexer/lexer.go` |
| Implement Go parser → AST | ✅ DONE | `compiler/internal/parser/parser.go` |
| Define AST types | ✅ DONE | `compiler/internal/parser/ast.go` |
| Basic type checker | ✅ DONE | `compiler/internal/checker/checker.go` |
| CLI skeleton | ✅ DONE | `compiler/cmd/dataprotoc/main.go` |
| Lexer tests | ✅ DONE | `compiler/internal/lexer/lexer_test.go` |

### Phase 2: Core Codegen
| Task | Status | File |
|------|--------|------|
| Common codegen utilities | ✅ DONE | `compiler/internal/codegen/codegen.go` |
| Proto generator | ✅ DONE | `compiler/internal/codegen/proto.go` |
| SQLite DDL generator | ✅ DONE | `compiler/internal/codegen/sql_sqlite.go` |
| Java mapper generator | ✅ DONE | `compiler/internal/codegen/java.go` |
| Java query builder generator | ✅ DONE | `compiler/internal/codegen/java.go` |

### Phase 3: Certification Infrastructure
| Task | Status | File |
|------|--------|------|
| Certificate format design | ✅ DONE | `runtime/java/.../Certificate.java` |
| Certificate authority | ⬜ TODO | `certification/ca/sign.go` |
| Runtime certificate validation | ✅ DONE | `runtime/java/.../Certificate.java` |
| Test suite framework | ⬜ TODO | `certification/test_suite/` |
| Principles documentation | ✅ DONE | `certification/principles.md` |

### Phase 4: Multi-Backend
| Task | Status | File |
|------|--------|------|
| Postgres DDL generator | ✅ DONE | `compiler/internal/codegen/sql_postgres.go` |
| MongoDB generator | ✅ DONE | `compiler/internal/codegen/mongodb.go` |
| Ceramic model generator | ⬜ TODO | `compiler/internal/codegen/ceramic.go` |
| Migration differ | ⬜ TODO | `compiler/internal/migration/differ.go` |
| Migration runner | ⬜ TODO | `runtime/*/migration.*` |

### Phase 5: Multi-Language Code Generation
| Task | Status | File |
|------|--------|------|
| Java codegen (server) | ✅ DONE | `compiler/internal/codegen/java.go` |
| Swift codegen (iOS client) | ✅ DONE | `compiler/internal/codegen/swift.go` |
| Kotlin codegen (Android client) | ✅ DONE | `compiler/internal/codegen/kotlin.go` |
| Python codegen | ✅ DONE | `compiler/internal/codegen/python.go` |
| Qt/C++ codegen | ✅ DONE | `compiler/internal/codegen/qt.go` |
| Java runtime | ✅ DONE | `runtime/java/` |
| Swift runtime | ⬜ TODO | `runtime/swift/` |
| Kotlin runtime | ⬜ TODO | `runtime/kotlin/` |
| Python runtime | ⬜ TODO | `runtime/python/` |
| Qt runtime | ⬜ TODO | `runtime/qt/` |

### Phase 6: Developer Experience
| Task | Status | File |
|------|--------|------|
| VS Code extension | ⬜ TODO | `tools/vscode-extension/` |
| LSP implementation | ⬜ TODO | `compiler/internal/lsp/` |
| Documentation site | ⬜ TODO | External |
| Example Aurora schemas | ✅ DONE | `examples/aurora/` |

---

## Architecture: Client/Server Roles

DataProto generates different code depending on the target's role:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   iOS (Swift)   │     │ Android (Kotlin)│     │   Other Clients │
│   gRPC Client   │     │   gRPC Client   │     │   gRPC Client   │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │ gRPC (port 50051)
                                 ▼
                    ┌─────────────────────────┐
                    │   Java Server (Aurora)  │
                    │   - SQL Repositories    │
                    │   - Direct DB Access    │
                    │   - Single Source Truth │
                    └────────────┬────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              ▼                  ▼                  ▼
         ┌─────────┐       ┌──────────┐       ┌─────────┐
         │ SQLite  │       │ Postgres │       │ MongoDB │
         │(device) │       │ (family) │       │  (doc)  │
         └─────────┘       └──────────┘       └─────────┘
```

| Language | Role | Generated Code |
|----------|------|----------------|
| **Java** | Server | SQL repositories, mappers, gRPC services |
| **Swift** | iOS Client | Data classes, proto mappers, gRPC client |
| **Kotlin** | Android Client | Data classes, proto mappers, gRPC client |
| **Python** | Scripts/Tools | Data classes, SQL repositories |
| **Qt/C++** | Desktop Client | QObject classes, SQL repositories |

**Key principle:** Only the Java server has direct database access. Mobile clients communicate via gRPC.

---

## Generated Code Examples

### Swift (from CalendarEvent entity)

**CalendarEvent.swift** - Entity struct:
```swift
public struct CalendarEvent: Codable, Identifiable, Sendable {
    public var id: String
    public var title: String
    public var startDate: Int64
    public var endDate: Int64?
    public var isAllDay: Bool = false
    // ...
}
```

**CalendarEventMapper.swift** - iOS native type mapping:
```swift
public extension CalendarEvent {
    init(ekEvent: EKEvent) {
        self.init(
            id: ekEvent.eventIdentifier,
            title: ekEvent.title ?? "",
            startDate: Int64(ekEvent.startDate.timeIntervalSince1970 * 1000),
            // ...
        )
    }

    func toProto() -> CalendarEventProto { ... }
}
```

### Python (from CalendarEvent entity)

**models.py** - Dataclass:
```python
@dataclass
class CalendarEvent:
    """Entity mapped to table 'calendar_events'."""

    id: str
    title: str
    start_date: int
    end_date: Optional[int] = None
    is_all_day: bool = False
```

**repositories.py** - Repository with queries:
```python
class CalendarEventRepository(BaseRepository):
    def events_by_date_range(self, after: int, before: int) -> List[CalendarEvent]:
        sql = "SELECT * FROM calendar_events WHERE start_date >= ? AND start_date < ? ORDER BY start_date ASC"
        # ...
```

### Kotlin (Android gRPC Client)

**CalendarEvent.kt** - Data class:
```kotlin
@Serializable
data class CalendarEvent(
    val id: String,
    val title: String,
    @SerialName("start_date")
    val startDate: Long,
    @SerialName("end_date")
    val endDate: Long? = null,
    @SerialName("is_all_day")
    val isAllDay: Boolean = false,
    // ...
)
```

**CalendarServiceClient.kt** - gRPC client with coroutines:
```kotlin
class CalendarServiceClient(
    private val host: String = "localhost",
    private val port: Int = 50051
) {
    private val stub: CalendarServiceGrpcKt.CalendarServiceCoroutineStub by lazy {
        CalendarServiceGrpcKt.CalendarServiceCoroutineStub(channel)
    }

    suspend fun pushEvents(items: List<CalendarEvent>): PushResult {
        return withContext(Dispatchers.IO) {
            val protos = CalendarEventMapper.toProtoList(items)
            stub.pushEvents(protos.asFlow())
        }
    }

    fun getEvents(request: GetEventsRequest): Flow<CalendarEvent> {
        return stub.getEvents(request).map { CalendarEventMapper.fromProto(it) }
    }
}
```

### Qt/C++ (from CalendarEvent entity)

**calendar_event.h** - QObject class (QML-ready):
```cpp
class CalendarEvent : public QObject
{
    Q_OBJECT
    QML_ELEMENT

    Q_PROPERTY(QString id READ id WRITE setId NOTIFY idChanged)
    Q_PROPERTY(QString title READ title WRITE setTitle NOTIFY titleChanged)
    Q_PROPERTY(qint64 startDate READ startDate WRITE setStartDate NOTIFY startDateChanged)
    // ...

public:
    explicit CalendarEvent(QObject *parent = nullptr);
    static CalendarEvent* create(QString id, QString title, qint64 startDate, ...);

    QString id() const;
    void setId(QString value);
    // ...

signals:
    void idChanged();
    void titleChanged();
    // ...
};
```

**calendar_event_repository.h** - Qt SQL repository:
```cpp
class CalendarEventRepository : public QObject
{
    Q_OBJECT

public:
    explicit CalendarEventRepository(QSqlDatabase db, QObject *parent = nullptr);

    void upsert(CalendarEvent *entity);
    CalendarEvent* findById(QString id, QObject *parent = nullptr);
    QList<CalendarEvent*> findAll(QObject *parent = nullptr);
    bool remove(QString id);

    // Generated from query definitions
    QList<CalendarEvent*> eventsByDateRange(qint64 after, qint64 before, QObject *parent = nullptr);
    QList<CalendarEvent*> upcomingEvents(int limit = 50, QObject *parent = nullptr);
};
```

**CMakeLists.txt** - Build configuration:
```cmake
find_package(Qt6 REQUIRED COMPONENTS Core Sql Qml)

add_library(acos_models ${SOURCES})
target_link_libraries(acos_models PRIVATE Qt6::Core Qt6::Sql Qt6::Qml)
```

---

## Language Specification

### Example DataProto File

```dataproto
package acos;

// Import other schemas
import "common.dataproto";

// Enum definition
enum MediaType {
  UNKNOWN = 0;
  IMAGE = 1;
  VIDEO = 2;
  LIVE_PHOTO = 3;
}

// Entity with storage annotations
@table("calendar_events")
@backends(sqlite, postgres)
entity CalendarEvent {
  // Primary key
  @pk id: string;

  // Required field with validation
  @required
  @length(1, 500)
  title: string;

  // Indexed timestamp field
  @indexed
  start_date: timestamp;

  // Optional field (nullable)
  end_date: timestamp?;

  // Boolean with default
  @default(false)
  is_all_day: bool;

  // String with max length
  @length(max: 2000)
  notes: string?;

  // Color as hex string
  @pattern("^#[0-9A-Fa-f]{6}$")
  calendar_color: string?;

  calendar_name: string?;
  location: string?;

  // Named query definitions
  query eventsByDateRange(after: timestamp, before: timestamp) {
    where start_date >= after AND start_date < before
    order_by start_date ASC
  }

  query upcomingEvents(limit: int32 = 10) {
    where start_date >= NOW()
    order_by start_date ASC
    limit limit
  }

  query searchEvents(term: string) {
    where title LIKE '%' || term || '%'
       OR notes LIKE '%' || term || '%'
    order_by start_date DESC
  }
}

// Service definition (generates gRPC service)
service CalendarService {
  rpc PushEvents(stream CalendarEvent) returns (PushResult);
  rpc GetEvents(GetEventsRequest) returns (stream CalendarEvent);
  rpc DeleteEvent(DeleteRequest) returns (Result);
}
```

### Type System

| DataProto Type | Proto Type | SQLite Type | Postgres Type | Java Type | Swift Type | Kotlin Type |
|----------------|------------|-------------|---------------|-----------|------------|-------------|
| `string` | `string` | `TEXT` | `TEXT` | `String` | `String` | `String` |
| `int32` | `int32` | `INTEGER` | `INTEGER` | `int` | `Int32` | `Int` |
| `int64` | `int64` | `INTEGER` | `BIGINT` | `long` | `Int64` | `Long` |
| `float` | `float` | `REAL` | `REAL` | `float` | `Float` | `Float` |
| `double` | `double` | `REAL` | `DOUBLE PRECISION` | `double` | `Double` | `Double` |
| `bool` | `bool` | `INTEGER` | `BOOLEAN` | `boolean` | `Bool` | `Boolean` |
| `bytes` | `bytes` | `BLOB` | `BYTEA` | `byte[]` | `Data` | `ByteArray` |
| `timestamp` | `int64` | `INTEGER` | `BIGINT` | `long` | `Int64` | `Long` |
| `T?` | `optional T` | nullable | nullable | `@Nullable T` | `T?` | `T?` |

### Annotations Reference

| Annotation | Target | Description |
|------------|--------|-------------|
| `@table("name")` | entity | SQL table name |
| `@backends(...)` | entity | Target storage backends |
| `@pk` | field | Primary key |
| `@required` | field | NOT NULL constraint |
| `@indexed` | field | Create index |
| `@unique` | field | Unique constraint |
| `@default(value)` | field | Default value |
| `@length(min, max)` | field | String length validation |
| `@pattern("regex")` | field | Regex validation |
| `@range(min, max)` | field | Numeric range validation |
| `@fk(Entity.field)` | field | Foreign key reference |

---

## Generated Output Examples

### From CalendarEvent Entity

**Generated Proto (`calendar.proto`):**
```protobuf
syntax = "proto3";
package acos;

message CalendarEvent {
  string id = 1;
  string title = 2;
  int64 start_date = 3;
  optional int64 end_date = 4;
  bool is_all_day = 5;
  optional string notes = 6;
  optional string calendar_color = 7;
  optional string calendar_name = 8;
  optional string location = 9;
}
```

**Generated SQLite DDL:**
```sql
CREATE TABLE IF NOT EXISTS calendar_events (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    start_date INTEGER NOT NULL,
    end_date INTEGER,
    is_all_day INTEGER DEFAULT 0,
    notes TEXT,
    calendar_color TEXT,
    calendar_name TEXT,
    location TEXT
);

CREATE INDEX IF NOT EXISTS idx_calendar_events_start_date
    ON calendar_events(start_date);
```

**Generated Java Repository:**
```java
public class CalendarEventRepository {
    private final DataProtoRuntime runtime;

    public CalendarEventRepository(DataProtoRuntime runtime) {
        runtime.requireCertified(); // HARD ENFORCEMENT
        this.runtime = runtime;
    }

    public void upsert(CalendarEvent event) {
        runtime.execute(
            "INSERT OR REPLACE INTO calendar_events " +
            "(id, title, start_date, end_date, is_all_day, notes, " +
            "calendar_color, calendar_name, location) " +
            "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
            event.getId(),
            event.getTitle(),
            event.getStartDate(),
            event.getEndDate(),
            event.getIsAllDay() ? 1 : 0,
            event.getNotes(),
            event.getCalendarColor(),
            event.getCalendarName(),
            event.getLocation()
        );
    }

    // Generated from: query eventsByDateRange(after, before)
    public List<CalendarEvent> eventsByDateRange(long after, long before) {
        return runtime.query(
            CalendarEvent.class,
            "SELECT * FROM calendar_events " +
            "WHERE start_date >= ? AND start_date < ? " +
            "ORDER BY start_date ASC",
            after, before
        );
    }

    // Generated from: query upcomingEvents(limit)
    public List<CalendarEvent> upcomingEvents(int limit) {
        return runtime.query(
            CalendarEvent.class,
            "SELECT * FROM calendar_events " +
            "WHERE start_date >= ? " +
            "ORDER BY start_date ASC LIMIT ?",
            System.currentTimeMillis(), limit
        );
    }
}
```

---

## Certification System

### Certificate Format

Certificates are signed JWTs containing:

```json
{
  "iss": "dataproto.aurora.dev",
  "sub": "app.example.com",
  "iat": 1704067200,
  "exp": 1735689600,
  "principles": {
    "data_sovereignty": true,
    "no_surveillance": true,
    "export_delete": true,
    "interoperability": true
  },
  "app_hash": "sha256:abc123...",
  "test_suite_version": "1.0.0"
}
```

### Principles

Certified implementations agree to:

1. **Data Sovereignty**: User data stored in user-controlled location
2. **No Surveillance**: No behavioral tracking or manipulation
3. **Export/Delete**: Users can export and delete their data
4. **Interoperability**: Works with Aurora ecosystem
5. **No Extraction**: Data stays on user's device/server

### Enforcement Points

```
Application Startup
       │
       ▼
┌──────────────────┐
│ Load Certificate │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐     ┌──────────────────┐
│ Verify Signature │────►│ INVALID: Refuse  │
│                  │ NO  │ to start         │
└────────┬─────────┘     └──────────────────┘
         │ YES
         ▼
┌──────────────────┐     ┌──────────────────┐
│ Check Expiration │────►│ EXPIRED: Refuse  │
│                  │ NO  │ to start         │
└────────┬─────────┘     └──────────────────┘
         │ YES
         ▼
┌──────────────────┐
│ Application Runs │
└──────────────────┘
```

---

## Directory Structure

```
dataproto/
├── IMPLEMENTATION.md          # This file
├── README.md                  # Project overview
├── LICENSE.md                 # Certification-required license
│
├── spec/
│   ├── grammar.ebnf           # Formal grammar
│   ├── type_system.md         # Type system documentation
│   └── annotations.md         # Annotation reference
│
├── compiler/                  # Go compiler (dataprotoc)
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── dataprotoc/
│   │       └── main.go        # CLI entry point
│   └── internal/
│       ├── lexer/
│       │   ├── lexer.go       # Tokenizer
│       │   ├── lexer_test.go
│       │   └── token.go       # Token definitions
│       ├── parser/
│       │   ├── parser.go      # Parser
│       │   ├── parser_test.go
│       │   └── ast.go         # AST node types
│       ├── checker/
│       │   ├── checker.go     # Type checker
│       │   └── checker_test.go
│       ├── ir/
│       │   └── ir.go          # Intermediate representation
│       └── codegen/
│           ├── codegen.go     # Common codegen utilities
│           ├── proto.go       # Proto file generator
│           ├── sql_sqlite.go  # SQLite DDL generator
│           ├── sql_postgres.go# Postgres DDL generator
│           ├── mongodb.go     # MongoDB JSON Schema + setup
│           ├── java.go        # Java codegen (server)
│           ├── kotlin.go      # Kotlin codegen (Android client)
│           ├── swift.go       # Swift codegen (iOS client)
│           ├── python.go      # Python codegen
│           └── qt.go          # Qt/C++ codegen
│
├── runtime/
│   ├── java/
│   │   ├── pom.xml
│   │   └── src/main/java/dev/dataproto/
│   │       ├── DataProtoRuntime.java
│   │       ├── Certificate.java
│   │       ├── CertificateValidator.java
│   │       ├── QueryExecutor.java
│   │       └── MigrationRunner.java
│   ├── swift/
│   │   ├── Package.swift
│   │   └── Sources/DataProto/
│   │       ├── DataProtoRuntime.swift
│   │       ├── Certificate.swift
│   │       └── QueryExecutor.swift
│   └── python/
│       ├── setup.py
│       └── dataproto/
│           ├── __init__.py
│           ├── runtime.py
│           ├── certificate.py
│           └── query.py
│
├── certification/
│   ├── principles.md          # What implementers agree to
│   ├── test_suite/
│   │   ├── crud.dataproto     # CRUD operation tests
│   │   ├── migrations.dataproto
│   │   └── transactions.dataproto
│   └── ca/
│       ├── sign.go            # Certificate signing
│       ├── verify.go          # Certificate verification
│       └── keys/              # (gitignored) CA keys
│
├── tools/
│   ├── vscode-extension/
│   │   ├── package.json
│   │   └── syntaxes/
│   │       └── dataproto.tmLanguage.json
│   └── intellij-plugin/
│
└── examples/
    └── aurora/
        ├── calendar.dataproto
        ├── reminders.dataproto
        └── photos.dataproto
```

---

## Build & Test Commands

```bash
# Build compiler
cd dataproto/compiler
go build -o bin/dataprotoc ./cmd/dataprotoc

# Run tests
go test ./...

# Compile a schema
./bin/dataprotoc compile \
  --input examples/aurora/calendar.dataproto \
  --output-dir generated/ \
  --targets proto,sqlite,java

# Verify existing schema matches
./bin/dataprotoc verify \
  --schema examples/aurora/calendar.dataproto \
  --proto ../Proto/acos/calendar_service.proto \
  --sql ../Java/apphandler/schema.sql
```

---

## Migration from Current Aurora Setup

### Step 1: Create Equivalent DataProto Schemas

Convert existing proto + SQL to DataProto:

| Current File | New DataProto |
|--------------|---------------|
| `Proto/acos/calendar_service.proto` + `DataHandler.java` SQL | `calendar.dataproto` |
| `Proto/acos/reminders_service.proto` + `DataHandler.java` SQL | `reminders.dataproto` |
| `Proto/acos/photos_service.proto` + `DataHandler.java` SQL | `photos.dataproto` |

### Step 2: Generate and Verify

```bash
# Generate from DataProto
dataprotoc compile --input calendar.dataproto --output generated/

# Diff against existing
diff generated/calendar.proto Proto/acos/calendar_service.proto
```

### Step 3: Replace Manual Code

Current `DataHandler.java`:
```java
// DELETE THIS - manually maintained
public CalendarEvent resultSetToCalendarEvent(ResultSet rs) { ... }
public void upsertCalendarEvent(CalendarEvent event) { ... }
```

Generated `CalendarEventRepository.java`:
```java
// USE THIS - auto-generated, type-safe, certified
public class CalendarEventRepository { ... }
```

---

## Bug Fixes History

### 2024-01-31: Initial Test Round

| Bug | Root Cause | Fix |
|-----|------------|-----|
| Unused variables in codegen | Leftover code from refactoring | Removed unused `colName`, `entityName`, `primaryKey` |
| Lexer number truncation | `pos` not updated before EOF check | Move `l.pos = l.readPos` before EOF check |
| Parser `limit` keyword conflict | `limit` used both as keyword and parameter name | Added `isKeywordAsIdent()` helper |
| Single quotes in examples | Lexer only supports double-quoted strings | Changed `'%'` to `"%"` in examples |
| Swift duplicate `id` property | Generated `id` computed property when PK already named `id` | Skip computed property when PK is `id` |
| Proto missing service types | Service references types like `PushResult` not generated | Added `collectSupportingTypes()` to auto-generate |

---

## Test Results

### Successful Compilation (2024-01-31)

All example schemas compile successfully:

```bash
$ ./dataprotoc compile --input calendar.dataproto --output-dir /tmp/test --targets all
Successfully compiled calendar.dataproto

Generated files:
- proto/acos.proto (with all supporting message types)
- sql/acos_schema.sql
- java/CalendarEventRepository.java, CalendarEventMapper.java
- swift/CalendarEvent.swift, CalendarEventMapper.swift, CalendarEventRepository.swift
- python/__init__.py, models.py, repositories.py, mappers.py
- qt/calendar_event.h, calendar_event.cpp, calendar_event_repository.h, calendar_event_repository.cpp, CMakeLists.txt
```

---

## Next Steps

**Completed:**
- [x] Create formal grammar → `spec/grammar.ebnf`
- [x] Implement lexer → `compiler/internal/lexer/`
- [x] Implement parser → `compiler/internal/parser/`
- [x] Build proto generator → generates valid proto3
- [x] Build SQLite generator → generates valid DDL
- [x] Build Java codegen → repositories and mappers
- [x] Build Swift codegen → with iOS native type mappers
- [x] Build Python codegen → dataclasses and repositories
- [x] Build Qt/C++ codegen → QObject classes for QML
- [x] Create example schemas → `examples/aurora/`

**Remaining:**

**High Priority (Data Integrity):**
1. **Validation Code Generator** → runtime checks for all languages
   - @required, @length, @pattern, @range enforcement
   - Generated for Java, Swift, Python, Qt
   - Reject bad data before it reaches DB

2. **Automated Test Generator** → verify data layer works
   - Validation tests (bad data rejected)
   - Repository tests (CRUD works)
   - Query tests (queries return correct data)
   - Fixtures (sample data for testing)

3. **Migration Differ** → additive-only schema evolution
   - Compare before/after schemas
   - Generate ALTER TABLE ADD COLUMN
   - Block destructive changes (delete/rename/type change)
   - Foundation columns are immutable

**Medium Priority (Completed):**
- [x] **Postgres DDL generator** → `sql_postgres.go` (family hubs)
- [x] **MongoDB generator** → JSON Schema + setup.js + Python repos
- [x] **Kotlin codegen** → Android gRPC client (like Swift for iOS)
- [x] **EventAttachment entity** → bytes fields for files/PDFs

**Future (Decentralized Layer):**
4. **CID Generation** → content addressing for P2P sync
5. **P2P Sync Protocol** → mesh network foundation

**Lower Priority:**
6. **Certificate Authority** → `certification/ca/sign.go`
7. **Swift/Kotlin/Python/Qt Runtimes** → certification checks
8. **VS Code Extension** → syntax highlighting + LSP

---

## Future: Decentralized Data Layer

DataProto can serve as the schema layer for a Ceramic-like P2P network, but with SQL instead of JSON:

```
DataProto Schema
      │
      ├──▶ SQLite (local storage, full SQL queries)
      ├──▶ Proto (wire format for sync)
      └──▶ CID (content addressing for P2P)
            │
            ▼
      P2P Sync (libp2p/IPFS)
            │
            ▼
      User Keys (DID) = ownership
```

Components to build:
- [ ] CID generator (row → hash → content ID)
- [ ] Change tracking (which rows changed since last sync)
- [ ] P2P discovery (find peers with same data)
- [ ] Conflict resolution (CR-SQLite integration)
- [ ] DID authentication (user keys sign changes)
