package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

type SyncState struct {
	TableName      string         `db:"table_name"`
	LastSyncTime   sql.NullTime   `db:"last_sync_time"`
	BinlogFile     sql.NullString `db:"binlog_file"`
	BinlogPosition sql.NullInt64  `db:"binlog_position"`
	RowsSynced     int64          `db:"rows_synced"`
	SyncDirection  string         `db:"sync_direction"`
	Status         string         `db:"status"`
	ErrorMessage   sql.NullString `db:"error_message"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

type Conflict struct {
	ID                 string         `db:"id"`
	TableName          string         `db:"table_name"`
	PrimaryKeyValue    string         `db:"primary_key_value"`
	LocalData          json.RawMessage `db:"local_data"`
	CloudData          json.RawMessage `db:"cloud_data"`
	ConflictType       string         `db:"conflict_type"`
	DetectedAt         time.Time      `db:"detected_at"`
	Resolved           bool           `db:"resolved"`
	ResolutionStrategy sql.NullString `db:"resolution_strategy"`
	ResolvedAt         sql.NullTime   `db:"resolved_at"`
	ResolvedData       json.RawMessage `db:"resolved_data"`
}

type SyncHistory struct {
	ID                string         `db:"id"`
	StartedAt         time.Time      `db:"started_at"`
	CompletedAt       sql.NullTime   `db:"completed_at"`
	Direction         string         `db:"direction"`
	TablesSynced      string         `db:"tables_synced"`
	TotalRows         int64          `db:"total_rows"`
	ConflictsDetected int            `db:"conflicts_detected"`
	Status            string         `db:"status"`
	ErrorMessage      sql.NullString `db:"error_message"`
}
