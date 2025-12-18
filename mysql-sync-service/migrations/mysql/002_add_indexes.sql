CREATE INDEX idx_conflicts_resolved ON conflicts(resolved);
CREATE INDEX idx_conflicts_table ON conflicts(table_name);
CREATE INDEX idx_history_started ON sync_history(started_at);
