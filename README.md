# Tactical Situation Display (TSD) — Backend Service

A high-performance, modular REST API service written in **Go (Golang)** for managing and persisting real-world and tactical entities with geospatial coordinates.

**Candidate:** Fredli Fourqoni  
**Email:** fredlifourqoni999@gmail.com  
**Position:** Software Developer  
**Company:** PT Len Innovation Technology (Len-IOT)

---

## 1. Overview & Features

This service serves as the core persistence and data validation engine for the Tactical Situation Display (TSD). It implements:
- **Full CRUD operations** for entities with geographical coordinates (\(\text{lat}, \text{lon}, \text{altitude}\)), bearing (\(\text{course}\)), speed, and tactical classification.
- **Strict Data Validation**: Symmetrical with frontend validation rules to ensure data integrity at API boundaries.
- **Auto Initial Seeding**: Automatically populates realistic tactical maritime contacts (e.g., naval vessels in Sunda Strait) on first startup.
- **Zero-CGO Pure Go SQLite Database**: Runs anywhere on Windows, Linux, and macOS without requiring GCC or MinGW compilers.
- **CORS Support**: Seamless integration with the frontend running on `http://localhost:8123`.
- **100% Colocated Automated Tests**: Complete test coverage across domain validators, SQLite repositories, services, and HTTP handlers.

---

## 2. Technology Stack & Decision Rationale

| Layer | Technology | Decision Rationale |
| --- | --- | --- |
| **Language** | **Go (Golang) 1.22+** | Selected for its strong static typing, high concurrent I/O performance, low memory footprint, and rapid cold-start times. |
| **Routing / HTTP** | **Standard Library `net/http`** | Leverages Go 1.22+ enhanced pattern routing (`POST /api/v1/entities/{id}`) without third-party framework overhead. |
| **Database** | **SQLite via `modernc.org/sqlite`** | 100% Pure Go SQLite implementation requiring **no CGO** and **no C compiler**. Embedded, zero-configuration, and durable. |
| **Concurrency Mode** | **WAL (Write-Ahead Logging)** | Configured with `_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)` for high-throughput concurrent reads and writes. |
| **Testing** | **Go `testing` + `httptest`** | Fast unit and integration tests without external test runner bloat. |

---

## 3. Installation & Setup

### Prerequisites
- [Go (Golang)](https://go.dev/dl/) version 1.22 or higher (`go version` to check).

### Running the Application

1. **Clone & Navigate to the Project:**
   ```powershell
   cd C:\Users\fredl\PROJECTS\leniot-intern\be-explore-maptalks
   ```

2. **Download Dependencies:**
   ```powershell
   go mod tidy
   ```

3. **Run Automated Test Suite:**
   ```powershell
   go test -v ./...
   ```

4. **Start the API Server:**
   ```powershell
   go run ./cmd/server/main.go
   ```

The server will initialize the SQLite database at `data/entities.db` and listen on:
- API Base URL: `http://localhost:8080`
- Health Check: `http://localhost:8080/api/v1/health`
- Entities Endpoint: `http://localhost:8080/api/v1/entities`

---

## 4. API Specification & Endpoints

### 4.1 Health Check
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

### 4.2 List Entities
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
        "created_at": "2026-11-03T03:20:00Z",
        "updated_at": "2026-11-03T03:20:00Z"
      }
    ]
  }
  ```

---

### 4.3 Get Entity by ID
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

### 4.4 Create Entity
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

### 4.5 Update Entity
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

### 4.6 Delete Entity
- **Endpoint:** `DELETE /api/v1/entities/{id}`
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "message": "Entity berhasil dihapus dari sistem"
  }
  ```

---

### 4.7 Reset / Re-seed Tactical Data
- **Endpoint:** `POST /api/v1/entities/seed`

---

## 5. Agentic AI Workflow

In compliance with the take-home test requirements, **Agentic AI** was leveraged purposefully across the software development lifecycle:

1. **Architectural Scaffolding & Design Alignment**:
   - Agentic AI analyzed the dual-repository structure and defined standard Go domain-driven package boundaries (`internal/entity`, `internal/database`, `internal/middleware`, `cmd/server`).
2. **Pure-Go Portability Optimization**:
   - The AI agent diagnosed that the Windows environment lacked a local GCC/MinGW compiler. Rather than introducing fragile external CGO dependencies (e.g. `mattn/go-sqlite3`), the agent selected `modernc.org/sqlite`, providing 100% pure Go compilation with native WAL support.
3. **Symmetrical Boundary Validation**:
   - Automated cross-referencing between frontend TypeScript validation utilities (`validation.ts`) and Go domain validators (`validator.go`) to guarantee zero validation drift.
4. **Colocated Automated Testing**:
   - The agent authored isolated unit and integration test suites (`validator_test.go`, `repository_test.go`, `service_test.go`, `handler_test.go`) utilizing temporary directories and `net/http/httptest` to achieve robust regression coverage.
5. **Quality Verification**:
   - The agent validated compilation, test execution, and Git identity integrity autonomously before submission.
