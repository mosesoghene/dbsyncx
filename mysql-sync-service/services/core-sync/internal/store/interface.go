package store

import (
	"context"
)

type Store interface {
	// Sync State
	GetSyncState(ctx context.Context, tableName string) (*SyncState, error)
	UpdateSyncState(ctx context.Context, state *SyncState) error
	
	// Conflicts
	CreateConflict(ctx context.Context, conflict *Conflict) error
	GetConflict(ctx context.Context, id string) (*Conflict, error)
	ListConflicts(ctx context.Context, resolved bool, limit, offset int) ([]*Conflict, error)
	ResolveConflict(ctx context.Context, id string, strategy string, resolvedData []byte) error
	
	// History
	CreateSyncHistory(ctx context.Context, history *SyncHistory) error
	UpdateSyncHistory(ctx context.Context, history *SyncHistory) error
	GetSyncHistory(ctx context.Context, limit, offset int) ([]*SyncHistory, error)
	
	// General
	Close() error
}
