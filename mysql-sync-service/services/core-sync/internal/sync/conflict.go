package sync

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"mysql-sync-service/internal/store"
)

type ConflictManager struct {
	store store.Store
}

func NewConflictManager(store store.Store) *ConflictManager {
	return &ConflictManager{
		store: store,
	}
}

func (cm *ConflictManager) DetectConflict(ctx context.Context, table string, pk string, localData, cloudData map[string]interface{}) (bool, *store.Conflict) {
	localHash := calculateHash(localData)
	cloudHash := calculateHash(cloudData)
	
	if localHash == cloudHash {
		return false, nil
	}
	
	// Conflict detected
	conflict := &store.Conflict{
		ID:              uuid.New().String(),
		TableName:       table,
		PrimaryKeyValue: pk,
		ConflictType:    "data_mismatch",
		DetectedAt:      time.Now(),
		Resolved:        false,
	}
	
	localBytes, _ := json.Marshal(localData)
	cloudBytes, _ := json.Marshal(cloudData)
	
	conflict.LocalData = json.RawMessage(localBytes)
	conflict.CloudData = json.RawMessage(cloudBytes)
	
	return true, conflict
}

func (cm *ConflictManager) RecordConflict(ctx context.Context, conflict *store.Conflict) error {
	return cm.store.CreateConflict(ctx, conflict)
}

func calculateHash(data map[string]interface{}) string {
	// TODO: Implement consistent hashing (sort keys, handle types)
	// For now, simple JSON string hash
	bytes, _ := json.Marshal(data)
	sum := sha256.Sum256(bytes)
	return fmt.Sprintf("%x", sum)
}

// Strategy interface for resolution
type ResolutionStrategy interface {
	Resolve(conflict *store.Conflict) (map[string]interface{}, error)
}

type LastWriteWinsStrategy struct {
	TimestampColumn string
}

func (s *LastWriteWinsStrategy) Resolve(conflict *store.Conflict) (map[string]interface{}, error) {
	// Parse local and cloud data to maps
	var local, cloud map[string]interface{}
	json.Unmarshal(conflict.LocalData, &local)
	json.Unmarshal(conflict.CloudData, &cloud)
	
	// Compare timestamps
	// ... logic to compare s.TimestampColumn
	
	return local, nil // Placeholder
}
