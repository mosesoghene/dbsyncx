package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	
	_ "github.com/go-sql-driver/mysql"
	"mysql-sync-service/internal/config"
	"mysql-sync-service/internal/logger"
	"go.uber.org/zap"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(cfg config.StateStorage) (*MySQLStore, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}
	
	// Retry loop for Ping
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		logger.Log.Info("Waiting for state DB...", zap.Error(err), zap.Int("attempt", i+1))
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to ping mysql after retries: %w", err)
	}
	
	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	
	return &MySQLStore{db: db}, nil
}

func (s *MySQLStore) Close() error {
	return s.db.Close()
}

func (s *MySQLStore) GetSyncState(ctx context.Context, tableName string) (*SyncState, error) {
	query := `SELECT table_name, last_sync_time, binlog_file, binlog_position, rows_synced, sync_direction, status, error_message, updated_at 
			  FROM sync_state WHERE table_name = ?`
	
	row := s.db.QueryRowContext(ctx, query, tableName)
	
	var state SyncState
	err := row.Scan(
		&state.TableName,
		&state.LastSyncTime,
		&state.BinlogFile,
		&state.BinlogPosition,
		&state.RowsSynced,
		&state.SyncDirection,
		&state.Status,
		&state.ErrorMessage,
		&state.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &state, nil
}

func (s *MySQLStore) UpdateSyncState(ctx context.Context, state *SyncState) error {
	query := `INSERT INTO sync_state (table_name, last_sync_time, binlog_file, binlog_position, rows_synced, sync_direction, status, error_message, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
			  ON DUPLICATE KEY UPDATE
			  last_sync_time = VALUES(last_sync_time),
			  binlog_file = VALUES(binlog_file),
			  binlog_position = VALUES(binlog_position),
			  rows_synced = VALUES(rows_synced),
			  sync_direction = VALUES(sync_direction),
			  status = VALUES(status),
			  error_message = VALUES(error_message),
			  updated_at = NOW()`
			  
	_, err := s.db.ExecContext(ctx, query,
		state.TableName,
		state.LastSyncTime,
		state.BinlogFile,
		state.BinlogPosition,
		state.RowsSynced,
		state.SyncDirection,
		state.Status,
		state.ErrorMessage,
	)
	
	return err
}

func (s *MySQLStore) CreateConflict(ctx context.Context, conflict *Conflict) error {
	query := `INSERT INTO conflicts (id, table_name, primary_key_value, local_data, cloud_data, conflict_type, detected_at, resolved)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
			  
	_, err := s.db.ExecContext(ctx, query,
		conflict.ID,
		conflict.TableName,
		conflict.PrimaryKeyValue,
		conflict.LocalData,
		conflict.CloudData,
		conflict.ConflictType,
		conflict.DetectedAt,
		conflict.Resolved,
	)
	
	return err
}

func (s *MySQLStore) GetConflict(ctx context.Context, id string) (*Conflict, error) {
	query := `SELECT id, table_name, primary_key_value, local_data, cloud_data, conflict_type, detected_at, resolved, resolution_strategy, resolved_at, resolved_data
			  FROM conflicts WHERE id = ?`
			  
	row := s.db.QueryRowContext(ctx, query, id)
	
	var c Conflict
	err := row.Scan(
		&c.ID,
		&c.TableName,
		&c.PrimaryKeyValue,
		&c.LocalData,
		&c.CloudData,
		&c.ConflictType,
		&c.DetectedAt,
		&c.Resolved,
		&c.ResolutionStrategy,
		&c.ResolvedAt,
		&c.ResolvedData,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &c, nil
}

func (s *MySQLStore) ListConflicts(ctx context.Context, resolved bool, limit, offset int) ([]*Conflict, error) {
	query := `SELECT id, table_name, primary_key_value, local_data, cloud_data, conflict_type, detected_at, resolved, resolution_strategy, resolved_at, resolved_data
			  FROM conflicts WHERE resolved = ? LIMIT ? OFFSET ?`
			  
	rows, err := s.db.QueryContext(ctx, query, resolved, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var conflicts []*Conflict
	for rows.Next() {
		var c Conflict
		err := rows.Scan(
			&c.ID,
			&c.TableName,
			&c.PrimaryKeyValue,
			&c.LocalData,
			&c.CloudData,
			&c.ConflictType,
			&c.DetectedAt,
			&c.Resolved,
			&c.ResolutionStrategy,
			&c.ResolvedAt,
			&c.ResolvedData,
		)
		if err != nil {
			return nil, err
		}
		conflicts = append(conflicts, &c)
	}
	
	return conflicts, nil
}

func (s *MySQLStore) ResolveConflict(ctx context.Context, id string, strategy string, resolvedData []byte) error {
	query := `UPDATE conflicts SET resolved = TRUE, resolution_strategy = ?, resolved_data = ?, resolved_at = NOW() WHERE id = ?`
	
	_, err := s.db.ExecContext(ctx, query, strategy, resolvedData, id)
	return err
}

func (s *MySQLStore) CreateSyncHistory(ctx context.Context, history *SyncHistory) error {
	query := `INSERT INTO sync_history (id, started_at, completed_at, direction, tables_synced, total_rows, conflicts_detected, status, error_message)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
			  
	_, err := s.db.ExecContext(ctx, query,
		history.ID,
		history.StartedAt,
		history.CompletedAt,
		history.Direction,
		history.TablesSynced,
		history.TotalRows,
		history.ConflictsDetected,
		history.Status,
		history.ErrorMessage,
	)
	
	return err
}

func (s *MySQLStore) UpdateSyncHistory(ctx context.Context, history *SyncHistory) error {
	query := `UPDATE sync_history SET completed_at = ?, total_rows = ?, conflicts_detected = ?, status = ?, error_message = ? WHERE id = ?`
	
	_, err := s.db.ExecContext(ctx, query,
		history.CompletedAt,
		history.TotalRows,
		history.ConflictsDetected,
		history.Status,
		history.ErrorMessage,
		history.ID,
	)
	
	return err
}

func (s *MySQLStore) GetSyncHistory(ctx context.Context, limit, offset int) ([]*SyncHistory, error) {
	query := `SELECT id, started_at, completed_at, direction, tables_synced, total_rows, conflicts_detected, status, error_message
			  FROM sync_history ORDER BY started_at DESC LIMIT ? OFFSET ?`
			  
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var history []*SyncHistory
	for rows.Next() {
		var h SyncHistory
		err := rows.Scan(
			&h.ID,
			&h.StartedAt,
			&h.CompletedAt,
			&h.Direction,
			&h.TablesSynced,
			&h.TotalRows,
			&h.ConflictsDetected,
			&h.Status,
			&h.ErrorMessage,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	
	return history, nil
}
