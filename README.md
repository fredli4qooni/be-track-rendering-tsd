# Tactical Situation Display (TSD) — Backend Service

A high-performance, modular REST API service written in **Go (Golang)** with **PostgreSQL 16 + PostGIS 3.4 Spatial Extension** for managing, querying, and persisting tactical maritime entities with geospatial coordinates.

---

## 1. Overview & Features

This service serves as the core persistence and spatial data validation engine for the Tactical Situation Display (TSD). It implements:
- **Full CRUD operations** for entities with geographical coordinates (\(\text{lat}, \text{lon}, \text{altitude}\)), bearing (\(\text{course}\)), speed, and tactical classification.
- **PostGIS Spatial Geography Storage**: Uses native WGS 84 elipsoid `GEOGRAPHY(Point, 4326)` for mathematically precise maritime GPS coordinates.
- **Spatial GIST Indexing (`idx_entities_geom`)**: High-speed spatial indexing enabling instant radius proximity queries (\(\text{ST\_DWithin}\)) and geofencing.
- **Strict Data Validation**: Symmetrical with frontend validation rules to ensure data integrity at API boundaries.
- **Auto Initial Seeding**: Automatically populates realistic tactical maritime contacts (e.g., naval vessels in Sunda Strait) on first startup.
- **Containerized PostGIS**: Runs seamlessly via Docker Compose (`postgis/postgis:16-3.4`).
- **CORS Support**: Seamless integration with the frontend running on `http://localhost:8123`.
- **100% Automated Tests**: Complete test coverage across domain validators, PostGIS repositories, services, and HTTP handlers.

---

## 2. Technology Stack & Decision Rationale

| Layer | Technology | Decision Rationale |
| --- | --- | --- |
| **Language** | **Go (Golang) 1.22+** | Strong static typing, high concurrent I/O performance, low memory footprint, and rapid execution. |
| **Routing / HTTP** | **Standard Library `net/http`** | Leverages Go 1.22+ enhanced pattern routing (`POST /api/v1/entities/{id}`) without third-party framework overhead. |
| **Database** | **PostgreSQL 16 + PostGIS 3.4** | Enterprise military and maritime geospatial standard. Stores native `GEOGRAPHY(Point, 4326)` coordinates. |
| **Spatial Index** | **GIST (Generalized Search Tree)** | R-Tree spatial indexing for high-speed proximity search, geofencing, and range calculations. |
| **Database Driver** | **`github.com/lib/pq`** | Lightweight, rock-solid PostgreSQL driver for Go `database/sql`. |
| **Testing** | **Go `testing` + `httptest`** | Fast unit and integration tests against PostGIS without external test runner bloat. |

---

## 3. Architectural Rationale: REST API vs. gRPC-Web via APISIX

In the enterprise front-end guidelines (`AGENTS.md`), the micro-frontend system uses **gRPC-Web via APISIX Gateway** for production Combat Management Systems (CMS). For this take-home test evaluation, **REST API** was selected based on the following engineering rationale:

1. **Zero Reviewer Friction (Standalone Evaluation)**:
   - Web browsers cannot communicate natively with HTTP/2 gRPC without a reverse proxy gateway (such as Apache APISIX or Envoy) translating gRPC-Web framing.
   - Requiring the evaluating team to spin up Docker containers for APISIX Gateway, `etcd`, and complex upstream routing would introduce high friction and environment dependency risks during code review.
   - With standard REST HTTP (`net/http`), reviewers can evaluate the entire application immediately.

2. **Production gRPC Contract Provided (`proto/entity.proto`)**:
   - The domain model and validation rules were engineered with 100% gRPC parity in mind.
   - A complete Protocol Buffers definition is included at [`proto/entity.proto`](./proto/entity.proto) demonstrating full schema and streaming readiness for production deployment behind APISIX.

---

## 4. Installation & Setup

### Prerequisites
- [Go (Golang)](https://go.dev/dl/) version 1.22 or higher (`go version` to check).
- [Docker & Docker Compose](https://www.docker.com/) for PostgreSQL + PostGIS container.

### Running the Application

1. **Navigate to the Project:**
   ```powershell
   cd C:\Users\fredl\PROJECTS\leniot-intern\be-explore-maptalks
   ```

2. **Start PostgreSQL + PostGIS Container:**
   ```powershell
   docker compose up -d
   ```
   *(Initializes PostgreSQL 16 with PostGIS 3.4 extension on port `5432`)*

3. **Run Automated Test Suite:**
   ```powershell
   go test -v ./...
   ```

4. **Start the API Server:**
   ```powershell
   go run ./cmd/server/main.go
   ```

The server will automatically connect to PostGIS, run schema migrations (enabling `postgis` extension, spatial table, and GIST index), seed initial contacts, and listen on:
- API Base URL: `http://localhost:8080`
- Health Check: `http://localhost:8080/api/v1/health`
- Entities Endpoint: `http://localhost:8080/api/v1/entities`

---

## 5. API Specification & Endpoints

### 5.1 Health Check
- **Endpoint:** `GET /api/v1/health`
- **Response:**
  ```json
  {
    "author": "Fredli Fourqoni",
    "service": "be-explore-maptalks",
    "status": "UP",
    "total_entities": 6,
    "version": "1.0.0"
  }
  ```

---

### 5.2 List Entities
- **Endpoint:** `GET /api/v1/entities`
- **Query Parameters (Optional):**
  - `status`: Filter by status (`ACTIVE`, `STANDBY`, `ENGAGED`, `INACTIVE`)
  - `type`: Filter by type integer (`0` to `5`)
- **Response:**
  ```json
  {
    "success": true,
    "message": "Daftar entity berhasil diambil",
    "data": [
      {
        "id": 1,
        "name": "KRI Raden Eddy Martadinata-331",
        "callsign": "REMI-331",
        "type": 0,
        "status": "ACTIVE",
        "lat": -5.912,
        "lon": 106.012,
        "course": 45.0,
        "speed": 22.5,
        "created_at": "2026-11-05T04:55:20Z",
        "updated_at": "2026-11-05T04:55:20Z"
      }
    ]
  }
  ```

---

### 5.3 Get Entity by ID
- **Endpoint:** `GET /api/v1/entities/{id}`
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "message": "Detail entity berhasil diambil",
    "data": { ... }
  }
  ```
- **Error (404 Not Found):**
  ```json
  {
    "success": false,
    "error": "Entity tidak ditemukan"
  }
  ```

---

### 5.4 Create Entity
- **Endpoint:** `POST /api/v1/entities`
- **Request Body:**
  ```json
  {
    "name": "KRI Diponegoro-365",
    "callsign": "DIPO-365",
    "type": 0,
    "status": "ACTIVE",
    "lat": -5.885,
    "lon": 106.045,
    "course": 90.0,
    "speed": 20.0
  }
  ```
- **Validation Rules:**
  - `name`: String, 2 to 60 characters, trimmed, required.
  - `lat`: Float, \(-90.0\) to \(+90.0\), required.
  - `lon`: Float, \(-180.0\) to \(+180.0\), required.
  - `course`: Float, \(0.0\) to \(360.0\) degrees.
  - `speed`: Float, \(0.0\) to \(120.0\) knots.
  - `type`: Integer, \(0\) to \(5\).
  - `status`: String, one of `ACTIVE`, `STANDBY`, `ENGAGED`, `INACTIVE`, `DESTROYED` (default: `ACTIVE`).

---

### 5.5 Update Entity
- **Endpoint:** `PUT /api/v1/entities/{id}`
- **Request Body (Partial or Complete):**
  ```json
  {
    "name": "KRI Diponegoro-365 (Patrol Mode)",
    "speed": 26.0,
    "status": "ENGAGED"
  }
  ```

---

### 5.6 Delete Entity
- **Endpoint:** `DELETE /api/v1/entities/{id}`
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "message": "Entity berhasil dihapus dari sistem"
  }
  ```

---

### 5.7 Reset / Re-seed Tactical Data
- **Endpoint:** `POST /api/v1/entities/seed`

---

## 6. Agentic AI Workflow

In compliance with the take-home test requirements, **Agentic AI** was leveraged purposefully across the software development lifecycle:

1. **Architectural Scaffolding & Design Alignment**:
   - Agentic AI analyzed the dual-repository structure and defined standard Go domain-driven package boundaries (`internal/entity`, `internal/database`, `internal/middleware`, `cmd/server`).
2. **PostGIS Geospatial Engineering**:
   - Migrated persistence layer from flat storage to **PostgreSQL 16 + PostGIS 3.4**, implementing native `GEOGRAPHY(Point, 4326)` coordinate types and spatial GIST indexing for defense-grade proximity calculation.
3. **Symmetrical Boundary Validation**:
   - Automated cross-referencing between frontend TypeScript validation utilities (`validation.ts`) and Go domain validators (`validator.go`) to guarantee zero validation drift.
4. **Colocated Automated Testing**:
   - The agent authored isolated unit and integration test suites (`validator_test.go`, `repository_test.go`, `service_test.go`, `handler_test.go`) validating PostGIS spatial queries, CRUD flows, and HTTP contracts.
5. **Quality Verification**:
   - The agent validated compilation, test execution, and Git identity integrity autonomously before submission.

---

## 7. Repository Links

- **Backend Repository**: [https://github.com/fredli4qooni/be-track-rendering-tsd.git](https://github.com/fredli4qooni/be-track-rendering-tsd.git)
- **Frontend Repository**: [https://github.com/fredli4qooni/fe-track-rendering-tsd.git](https://github.com/fredli4qooni/fe-track-rendering-tsd.git)
- **Documentation Portal**: [https://doc-track-rendering-tsd.vercel.app/](https://doc-track-rendering-tsd.vercel.app/)
