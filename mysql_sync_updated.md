# MySQL Database Sync Service - Implementation Plan

## Project Overview

Build a high-performance microservice in Go that synchronizes MySQL databases between local and cloud instances using binlog replication for real-time change detection. The system is fully containerized and follows microservices architecture with clear separation of concerns.

---

## Senior Engineer Instructions

**You are a senior Go engineer with 10+ years of experience building distributed systems.**

After reading this complete specification, you must:

1. **Analyze the requirements** and identify any potential issues, edge cases, or architectural concerns
2. **Break down the implementation** into granular, actionable tasks with clear acceptance criteria
3. **Create a task list** that includes:
   - Setup and infrastructure tasks
   - Core feature implementation tasks
   - Testing and validation tasks
   - Documentation tasks
   - Performance optimization tasks
4. **Identify dependencies** between tasks and suggest an optimal implementation order
5. **Flag any ambiguities** or missing information that needs clarification
6. **Propose technical decisions** where the spec leaves room for interpretation
7. **Consider operational concerns** like monitoring, logging, error recovery, and deployment
8. **Think about scalability** and how the system will behave under load

Your task breakdown should be detailed enough that a mid-level engineer could implement each task independently. Include:
- Task title and description
- Acceptance criteria
- Key technical considerations
- Potential pitfalls to avoid
- Testing approach
- Estimated complexity (S/M/L/XL)

---

## Tech Stack

**Framework:** Chi Router (lightweight, idiomatic Go HTTP router)

**Core Dependencies:**
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/go-sql-driver/mysql` - MySQL driver
- `github.com/go-mysql-org/go-mysql` - Binlog replication
- `github.com/robfig/cron/v3` - Scheduler
- `go.uber.org/zap` - Structured logging
- `github.com/spf13/viper` - Configuration management

**State Storage:**
- MySQL or SQLite (configurable)
- Used for sync state, binlog positions, conflict tracking, and sync history

**Containerization:**
- Docker for all services
- Docker Compose for local development
- Kubernetes-ready deployment manifests

---

## Architecture Requirements

### Microservices Architecture

The system consists of **three independent services**:

1. **Core Sync Service** (Port 8080)
   - Binlog listener
   - Worker pool
   - Conflict detection
   - Sync manager
   - REST API

2. **State Storage Service**
   - MySQL or SQLite database
   - Stores sync state, conflicts, history
   - Accessed only by Core Sync Service

3. **Web UI Service** (Port 3000)
   - Standalone web application
   - React/Vue/vanilla JavaScript frontend
   - Communicates with Core Sync Service via REST API only
   - No direct database access
   - No shared libraries with Core Sync Service

**Service Communication:**
- Web UI → Core Sync Service: REST API + SSE
- Core Sync Service → State Storage: Database queries
- **No direct coupling** between Web UI and Core Sync Service

---

## Core Components (Sync Service)

### 1. Binlog Listener

**Responsibilities:**
- Connect to MySQL as replication client
- Stream binlog events in real-time
- Parse INSERT/UPDATE/DELETE operations
- Filter events for configured tables only
- Push changes to buffered queue (10K capacity)
- Store binlog position in state database after each batch
- Resume from last position on restart

**Implementation Requirements:**
- Use descriptive variable names: `binlogEvent`, `tableConfig`, `changeQueue`
- Handle connection failures with exponential backoff
- Log all events with structured logging
- Graceful shutdown with context cancellation

### 2. Worker Pool

**Responsibilities:**
- Configurable number of concurrent workers (default: 8)
- Each worker consumes from change queue
- Batch changes (1000 rows per batch)
- Execute batch INSERT/UPDATE/DELETE in transactions
- Use prepared statements for performance
- Log progress and errors with structured logging

**Implementation Requirements:**
- Descriptive worker identifiers: `workerID`, `batchProcessor`
- Proper synchronization with sync.WaitGroup
- Transaction rollback on errors
- Resource cleanup with defer statements

### 3. Conflict Detection & Resolution

**Responsibilities:**
- Detect conflicts by comparing row hashes (exclude timestamps)
- Support three resolution strategies:
  - Last-write-wins: Compare timestamp columns
  - Source priority: Always prefer local or cloud
  - Manual: Store conflict for user review
- Store unresolved conflicts in state database
- Expose conflicts via REST API

**Implementation Requirements:**
- Clear strategy pattern implementation
- Descriptive method names: `detectConflict()`, `resolveByTimestamp()`
- Comprehensive conflict metadata storage
- Thread-safe conflict handling

### 4. Sync Manager

**Responsibilities:**
- Orchestrate sync operations
- Manage bidirectional sync flow
- Track sync status and progress
- Publish real-time progress events
- Handle sync start/stop/status requests
- Maintain sync history in state database

**Implementation Requirements:**
- State machine for sync lifecycle: idle → running → completed/failed
- Prevent concurrent syncs with mutex or atomic flags
- Progress tracking with percentages and row counts
- SSE publisher for real-time updates

### 5. REST API (Chi Router)

**Endpoints:**
- `POST /api/v1/sync/trigger` - Start manual sync
- `POST /api/v1/sync/stop` - Cancel running sync
- `GET /api/v1/sync/status` - Current sync state
- `GET /api/v1/sync/history` - Past sync runs
- `GET /api/v1/sync/stream` - SSE for real-time progress
- `GET /api/v1/conflicts` - List unresolved conflicts
- `POST /api/v1/conflicts/:id/resolve` - Resolve conflict
- `GET /api/v1/config` - Get current configuration
- `PUT /api/v1/config` - Update configuration
- `GET /api/v1/metrics` - Prometheus-style metrics
- `GET /health` - Health check (no auth)

**Implementation Requirements:**
- Bearer token authentication middleware
- Skip auth for `/health` endpoint
- Return proper HTTP status codes (401, 403, 500, etc.)
- Request/response logging
- CORS headers for Web UI service

### 6. State Storage Layer

**Database Schema:**

```sql
-- Sync state per table
CREATE TABLE sync_state (
    table_name VARCHAR(255) PRIMARY KEY,
    last_sync_time TIMESTAMP,
    binlog_file VARCHAR(255),
    binlog_position BIGINT,
    rows_synced BIGINT,
    sync_direction VARCHAR(50),
    status VARCHAR(50),
    error_message TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Unresolved conflicts
CREATE TABLE conflicts (
    id VARCHAR(36) PRIMARY KEY,
    table_name VARCHAR(255),
    primary_key_value VARCHAR(255),
    local_data JSON,
    cloud_data JSON,
    conflict_type VARCHAR(50),
    detected_at TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE,
    resolution_strategy VARCHAR(50),
    resolved_at TIMESTAMP,
    resolved_data JSON
);

-- Sync history
CREATE TABLE sync_history (
    id VARCHAR(36) PRIMARY KEY,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    direction VARCHAR(50),
    tables_synced TEXT,
    total_rows BIGINT,
    conflicts_detected INT,
    status VARCHAR(50),
    error_message TEXT
);
```

**Implementation Requirements:**
- Support both MySQL and SQLite with interface abstraction
- Connection pooling configuration
- Prepared statements for all queries
- Proper transaction handling
- Migration system for schema updates

### 7. Scheduler

**Responsibilities:**
- Use cron syntax for intervals
- Fallback sync when binlog streaming fails
- Configurable enable/disable
- Don't trigger if manual sync is running

**Implementation Requirements:**
- Use robfig/cron library
- Descriptive job names: `scheduledSyncJob`
- Check sync status before triggering
- Log all scheduled executions

---

## Web UI Service (Separate Container)

### Architecture

**Technology Stack:**
- Frontend: React, Vue, or vanilla JavaScript (your choice)
- Build tool: Vite or Webpack
- HTTP server: Nginx or Node.js serve
- Port: 3000

**Features:**
- Single-page dashboard
- Start/Stop sync buttons
- Real-time progress display (connect to SSE endpoint)
- Conflict list with resolve buttons
- Recent sync history table
- Status indicator (idle/syncing/error)
- Configuration viewer/editor

**Communication:**
- All data fetched via REST API from Core Sync Service
- Real-time updates via SSE connection
- Bearer token authentication
- No direct database access
- No shared code with Core Sync Service

**Docker Requirements:**
- Separate Dockerfile
- Static file serving
- Environment variable for API URL
- CORS handling on backend

---

## Configuration Structure (YAML)

```yaml
databases:
  local:
    host: localhost
    port: 3306
    user: sync_user
    password: password
    database: myapp_local
    replication_user: repl_user
    replication_password: repl_password
  
  cloud:
    host: db.example.com
    port: 3306
    user: sync_user
    password: password
    database: myapp_cloud
    replication_user: repl_user
    replication_password: repl_password

state_storage:
  type: mysql  # or sqlite
  host: state-db
  port: 3306
  user: state_user
  password: state_password
  database: sync_state
  # For SQLite:
  # file_path: ./data/sync_state.db

sync:
  mode: bidirectional  # local_to_cloud | cloud_to_local | bidirectional
  
  tables:
    - name: users
      conflict_resolution: last_write_wins
      batch_size: 5000
      primary_key: id
      timestamp_column: updated_at
      
    - name: orders
      conflict_resolution: manual
      batch_size: 10000
      primary_key: order_id
      timestamp_column: modified_at
  
  workers: 8
  realtime: true
  batch_insert_size: 1000
  
scheduler:
  enabled: true
  interval: "*/10 * * * *"  # Every 10 minutes

server:
  port: 8080
  host: 0.0.0.0
  auth_token: "your-secret-token"
  read_timeout: 30s
  write_timeout: 30s
  cors_origins:
    - "http://localhost:3000"
    - "https://sync-ui.example.com"

logging:
  level: info
  format: json
```

---

## Directory Structure

```
mysql-sync-service/
├── services/
│   ├── core-sync/
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── api/
│   │   │   │   ├── handlers.go
│   │   │   │   ├── middleware.go
│   │   │   │   └── routes.go
│   │   │   ├── sync/
│   │   │   │   ├── manager.go
│   │   │   │   ├── worker.go
│   │   │   │   ├── binlog.go
│   │   │   │   ├── conflict.go
│   │   │   │   └── strategy.go
│   │   │   ├── store/
│   │   │   │   ├── interface.go
│   │   │   │   ├── mysql.go
│   │   │   │   ├── sqlite.go
│   │   │   │   └── models.go
│   │   │   ├── config/
│   │   │   │   ├── config.go
│   │   │   │   └── loader.go
│   │   │   └── database/
│   │   │       ├── mysql.go
│   │   │       └── queries.go
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── web-ui/
│       ├── src/
│       │   ├── components/
│       │   │   ├── Dashboard.js
│       │   │   ├── SyncControls.js
│       │   │   ├── ConflictList.js
│       │   │   ├── SyncHistory.js
│       │   │   └── StatusIndicator.js
│       │   ├── services/
│       │   │   └── api.js
│       │   ├── App.js
│       │   └── index.js
│       ├── public/
│       │   └── index.html
│       ├── Dockerfile
│       ├── nginx.conf
│       ├── package.json
│       └── vite.config.js
│
├── deployments/
│   ├── docker-compose.yml
│   ├── docker-compose.dev.yml
│   └── kubernetes/
│       ├── core-sync-deployment.yaml
│       ├── web-ui-deployment.yaml
│       ├── state-db-deployment.yaml
│       └── ingress.yaml
│
├── migrations/
│   ├── mysql/
│   │   ├── 001_initial_schema.sql
│   │   └── 002_add_indexes.sql
│   └── sqlite/
│       ├── 001_initial_schema.sql
│       └── 002_add_indexes.sql
│
├── config.yaml
├── config.example.yaml
├── Makefile
└── README.md
```

---

## Docker Containerization

### 1. Core Sync Service Dockerfile

```dockerfile
# Multi-stage build for smaller image
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY services/core-sync/go.mod services/core-sync/go.sum ./
RUN go mod download

# Copy source code
COPY services/core-sync/ ./

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sync-service ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/sync-service .

# Copy migrations
COPY migrations/ ./migrations/

EXPOSE 8080

CMD ["./sync-service"]
```

### 2. Web UI Service Dockerfile

```dockerfile
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

COPY services/web-ui/package*.json ./
RUN npm ci

COPY services/web-ui/ ./
RUN npm run build

# Production stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY services/web-ui/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
```

### 3. Docker Compose

```yaml
version: '3.8'

services:
  state-db:
    image: mysql:8.0
    container_name: sync-state-db
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: sync_state
      MYSQL_USER: state_user
      MYSQL_PASSWORD: state_password
    volumes:
      - state-db-data:/var/lib/mysql
      - ./migrations/mysql:/docker-entrypoint-initdb.d
    ports:
      - "3307:3306"
    networks:
      - sync-network

  core-sync:
    build:
      context: .
      dockerfile: services/core-sync/Dockerfile
    container_name: core-sync-service
    depends_on:
      - state-db
    environment:
      - CONFIG_PATH=/app/config.yaml
    volumes:
      - ./config.yaml:/app/config.yaml
    ports:
      - "8080:8080"
    networks:
      - sync-network
    restart: unless-stopped

  web-ui:
    build:
      context: .
      dockerfile: services/web-ui/Dockerfile
    container_name: sync-web-ui
    depends_on:
      - core-sync
    environment:
      - API_URL=http://core-sync:8080
    ports:
      - "3000:3000"
    networks:
      - sync-network
    restart: unless-stopped

networks:
  sync-network:
    driver: bridge

volumes:
  state-db-data:
```

---

## Critical Code Requirements

### Variable Naming Standards

**MANDATORY RULES:**
- ALL variable names MUST be descriptive and readable
- NO single-letter variables (no i, j, x, y, etc.)
- Use full words in camelCase
- Be specific about what the variable represents

**Examples:**

❌ **BAD:**
```go
for i := 0; i < len(rows); i++ {
    r := rows[i]
    e := process(r)
    if e != nil {
        return e
    }
}
```

✅ **GOOD:**
```go
for rowIndex := 0; rowIndex < len(tableRows); rowIndex++ {
    currentRow := tableRows[rowIndex]
    processingError := processTableRow(currentRow)
    if processingError != nil {
        return processingError
    }
}
```

**Acceptable Conventions:**
- `ctx` for context.Context
- `err` for error returns
- `db` for database connections
- `wg` for sync.WaitGroup
- Standard Go library conventions

**Domain-Specific Names:**
- `binlogEvent` not `event`
- `conflictStrategy` not `strategy`
- `syncWorker` not `worker`
- `tableConfig` not `config`
- `changeQueue` not `queue`

### Error Handling

**Requirements:**
- Check EVERY error return value
- Log errors with context using zap
- Include relevant data in error logs: table name, row count, operation
- Don't panic - return errors up the call stack
- Use wrapped errors for context: `fmt.Errorf("failed to sync table %s: %w", tableName, err)`

**Example:**
```go
connectionResult, connectionError := database.Connect(databaseConfig)
if connectionError != nil {
    logger.Error("failed to connect to database",
        zap.String("host", databaseConfig.Host),
        zap.Int("port", databaseConfig.Port),
        zap.Error(connectionError),
    )
    return fmt.Errorf("database connection failed: %w", connectionError)
}
```

### Performance Requirements

- Use connection pooling (set MaxOpenConns and MaxIdleConns)
- Batch database operations (minimum 1000 rows per batch)
- Use prepared statements for repeated queries
- Stream large result sets, don't load everything into memory
- Close resources properly with defer statements
- Use context for timeouts and cancellation

### Concurrency Requirements

- Use context.Context for cancellation propagation
- Protect shared state with sync.Mutex or sync.RWMutex
- Use buffered channels with appropriate sizes
- Wait for goroutines with sync.WaitGroup
- Avoid goroutine leaks with proper cleanup

---

## Implementation Tasks (To Be Completed by Senior Engineer)

As a senior engineer, after reviewing this specification, you must:

1. **Create a detailed task breakdown** with:
   - Specific implementation tasks
   - Acceptance criteria for each task
   - Dependencies between tasks
   - Estimated complexity

2. **Identify technical decisions** needed:
   - Specific libraries or approaches
   - Performance optimization strategies
   - Testing strategies
   - Deployment considerations

3. **Flag any issues**:
   - Missing requirements
   - Potential bottlenecks
   - Security concerns
   - Scalability limitations

4. **Propose architecture improvements**:
   - Better approaches if you see them
   - Additional components that might be needed
   - Observability enhancements

5. **Create implementation order**:
   - Which tasks should be done first
   - What can be parallelized
   - Critical path identification

Your output should be a comprehensive task list that a team can use to implement this system.

---

## Success Criteria

### Performance Targets
- Sync 100K rows in under 10 seconds
- Real-time lag under 100ms from change to sync
- Memory usage under 200MB under load
- Support 20+ concurrent table syncs

### Reliability
- Recover from network interruptions
- Resume from last binlog position on restart
- No data loss on service crash
- Transactional consistency (all or nothing per batch)

### Observability
- Structured JSON logs with trace IDs
- Real-time progress via SSE
- Metrics for sync duration, row counts, error rates
- Health endpoint for monitoring
- Prometheus-compatible metrics

### Operational
- All services run in Docker containers
- Docker Compose for local development
- Kubernetes-ready deployments
- Configuration via environment variables or config files
- Database migrations handled automatically
- Graceful shutdown handling

---

## Development Approach

### Phase 1: Foundation (Week 1)
1. Set up project structure
2. Configure Docker containers
3. Implement configuration management
4. Set up state database with migrations
5. Establish logging and monitoring

### Phase 2: Core Sync Engine (Week 2-3)
1. Database connection management
2. Binlog listener implementation
3. Worker pool and change processing
4. Basic sync manager (one direction)

### Phase 3: Advanced Features (Week 4)
1. Bidirectional sync
2. Conflict detection and resolution
3. Scheduler implementation

### Phase 4: API & UI (Week 5)
1. REST API implementation
2. SSE for real-time updates
3. Web UI development
4. Integration testing

### Phase 5: Production Readiness (Week 6)
1. Performance testing and optimization
2. Error recovery testing
3. Documentation
4. Deployment guides

---

## Notes

- MySQL replication user needs REPLICATION SLAVE and REPLICATION CLIENT privileges
- Binlog format must be ROW (not STATEMENT or MIXED)
- Test with MySQL 5.7+ or 8.0+
- Consider table size - tables over 10M rows may need special handling
- Conflict resolution requires timestamp columns to exist
- For tables without timestamps, only source_priority or manual resolution work
- Web UI must be completely decoupled from Core Sync Service
- All services must be independently deployable and scalable
- State database can be MySQL or SQLite based on deployment needs
- Use descriptive variable names throughout the codebase - no exceptions