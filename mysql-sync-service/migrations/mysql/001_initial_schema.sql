-- Sync state per table
CREATE TABLE IF NOT EXISTS sync_state (
    table_name VARCHAR(255) PRIMARY KEY,
    last_sync_time TIMESTAMP NULL,
    binlog_file VARCHAR(255),
    binlog_position BIGINT,
    rows_synced BIGINT DEFAULT 0,
    sync_direction VARCHAR(50),
    status VARCHAR(50),
    error_message TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Unresolved conflicts
CREATE TABLE IF NOT EXISTS conflicts (
    id VARCHAR(36) PRIMARY KEY,
    table_name VARCHAR(255),
    primary_key_value VARCHAR(255),
    local_data JSON,
    cloud_data JSON,
    conflict_type VARCHAR(50),
    detected_at TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE,
    resolution_strategy VARCHAR(50),
    resolved_at TIMESTAMP NULL,
    resolved_data JSON
);

-- Sync history
CREATE TABLE IF NOT EXISTS sync_history (
    id VARCHAR(36) PRIMARY KEY,
    started_at TIMESTAMP,
    completed_at TIMESTAMP NULL,
    direction VARCHAR(50),
    tables_synced TEXT,
    total_rows BIGINT DEFAULT 0,
    conflicts_detected INT DEFAULT 0,
    status VARCHAR(50),
    error_message TEXT
);
