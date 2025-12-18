package sync

import (
	"context"
	"fmt"
	"sync"
	
	"mysql-sync-service/internal/config"
	"mysql-sync-service/internal/database"
	"mysql-sync-service/internal/logger"
	"mysql-sync-service/internal/store"
)

type Manager struct {
	cfg            *config.Config
	localDB        *database.Database
	cloudDB        *database.Database
	store          store.Store
	binlogListener *BinlogListener
	workerPool     *WorkerPool
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.Mutex
	status         string
}

func NewManager(cfg *config.Config, store store.Store) (*Manager, error) {
	// Connect to local DB
	localDB, err := database.NewDatabase(cfg.Databases.Local)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to local db: %w", err)
	}

	// Connect to cloud DB
	cloudDB, err := database.NewDatabase(cfg.Databases.Cloud)
	if err != nil {
		localDB.Close()
		return nil, fmt.Errorf("failed to connect to cloud db: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		cfg:     cfg,
		localDB: localDB,
		cloudDB: cloudDB,
		store:   store,
		ctx:     ctx,
		cancel:  cancel,
		status:  "idle",
	}, nil
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status == "running" {
		return fmt.Errorf("sync is already running")
	}

	logger.Log.Info("Starting sync manager")

	// Initialize Binlog Listener
	// Determine source based on config. For bidirectional, we might need two listeners.
	// For simplicity, let's assume Local -> Cloud for now as per "Phase 2"
	
	listener, err := NewBinlogListener(m.cfg.Databases.Local, m.cfg.Sync.Tables)
	if err != nil {
		return err
	}
	m.binlogListener = listener

	// Initialize Worker Pool (target is Cloud)
	m.workerPool = NewWorkerPool(m.cfg.Sync, m.cloudDB, m.store, listener.Events())
	m.workerPool.Start()

	// Start Listener
	if err := m.binlogListener.Start(); err != nil {
		m.workerPool.Stop()
		return err
	}

	m.status = "running"
	return nil
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status != "running" {
		return
	}

	logger.Log.Info("Stopping sync manager")

	if m.binlogListener != nil {
		m.binlogListener.Stop()
	}

	if m.workerPool != nil {
		m.workerPool.Stop()
	}

	m.status = "idle"
}

func (m *Manager) Close() {
	m.Stop()
	m.localDB.Close()
	m.cloudDB.Close()
}

func (m *Manager) GetStatus() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}
