# AGENT.md — Backend Tactical Situation Display (TSD)

Instructions for AI coding agents and human developers maintaining the Go backend service.

---

## 1. System Architecture & Boundaries

The backend service is structured using clean, idiomatic Go layered architecture without external heavyweight frameworks:

```
be-explore-maptalks/
├── cmd/
│   └── server/
│       └── main.go          # Application entrypoint & graceful shutdown
├── internal/
│   ├── database/
│   │   └── postgres.go      # PostgreSQL + PostGIS initialization & spatial auto-migration
│   ├── entity/
│   │   ├── model.go         # Domain models, enums & API response structs
│   │   ├── validator.go     # Strict input boundary validation (matches frontend rules)
│   │   ├── validator_test.go# Unit tests for domain validations
│   │   ├── repository.go    # Data access layer & PostGIS spatial queries (ST_MakePoint, ST_DWithin)
│   │   ├── repository_test.go# PostGIS spatial and CRUD integration tests
│   │   ├── service.go       # Core business logic & default seed generator
│   │   ├── service_test.go  # Business logic tests
│   │   ├── handler.go       # HTTP handlers (Go 1.22+ native pattern routing)
│   │   └── handler_test.go  # HTTP integration tests via net/http/httptest
│   └── middleware/
│       ├── cors.go          # CORS middleware for frontend communication
│       └── logger.go        # HTTP request structured logging
├── docker-compose.yml       # PostgreSQL 16 + PostGIS 3.4 container configuration
├── AGENT.md                 # Agentic AI guidelines & architecture
├── go.mod
├── go.sum
└── README.md                # Setup & API documentation
```

---

## 2. Technology Stack & Decision Rationale

| Concern | Choice | Rationale |
| --- | --- | --- |
| **Language** | Go (Golang) 1.22+ | Statically typed, high throughput, low memory footprint, ideal for real-time tactical streaming. |
| **HTTP Routing** | Standard library `net/http` | Native pattern routing (`"GET /api/v1/entities/{id}"`) without bloated third-party dependencies. |
| **Database** | PostgreSQL 16 + PostGIS 3.4 | Enterprise military & maritime standard for geospatial data; native `GEOGRAPHY(Point, 4326)`. |
| **Spatial Index** | GIST (Generalized Search Tree) | High-speed R-Tree proximity search and geofencing calculation. |
| **Database Driver** | `github.com/lib/pq` | Lightweight, battle-tested PostgreSQL driver for Go `database/sql`. |
| **Testing** | Standard `testing` + `httptest` | Fast, colocated unit and integration tests with zero external test runners. |
| **API Protocol** | REST JSON + Protobuf Contract | Zero-friction evaluation; includes `proto/entity.proto` for enterprise gRPC-Web parity. |

---

### Architectural Note: REST vs. gRPC-Web via APISIX
While production micro-frontends specify gRPC-Web via APISIX, a pure Go REST API was chosen for the standalone assessment to eliminate external proxy (APISIX/Envoy) dependencies during evaluation, while providing full schema equivalence through `proto/entity.proto`.

---

## 3. Coding Conventions & Quality Gates

1. **Strict Error Handling**: Never ignore errors. Always return wrapped errors with context: `fmt.Errorf("action context: %w", err)`.
2. **Input Validation Integrity**: 
   - Coordinates: Latitude must be strictly \([-90, +90]\), Longitude \([-180, +180]\).
   - Course: \([0, 360]\) degrees.
   - Speed: \([0, 120]\) knots.
   - Names: \([2, 60]\) characters, whitespace-trimmed.
3. **Colocated Testing**: Every package must maintain colocated `*_test.go` files.

---

## 4. Definition of Done (DoD)

Before declaring any change complete:
- [ ] `go build ./...` compiles cleanly with 0 errors.
- [ ] `go test -v ./...` executes and passes all test suites.
- [ ] New functionality includes colocated `*_test.go` covering edge cases.
- [ ] Documentation (`README.md`, `AGENT.md`) reflects any API changes.
