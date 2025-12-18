package sync

import (
	"context"
	"fmt"

	"github.com/go-mysql-org/go-mysql/canal"
	"go.uber.org/zap"

	"mysql-sync-service/internal/config"
	"mysql-sync-service/internal/logger"
)

type BinlogListener struct {
	cfg        config.DatabaseConnection
	canal      *canal.Canal
	eventChan  chan BinlogEvent
	ctx        context.Context
	cancel     context.CancelFunc
	tables     map[string]bool // Whitelist of tables
}

func NewBinlogListener(cfg config.DatabaseConnection, tables []config.TableConfig) (*BinlogListener, error) {
	tableMap := make(map[string]bool)
	var tableRegex []string
	for _, t := range tables {
		tableMap[t.Name] = true
		tableRegex = append(tableRegex, fmt.Sprintf("^%s\\.%s$", cfg.Database, t.Name))
	}

	c, err := canal.NewCanal(&canal.Config{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:     cfg.ReplicationUser,
		Password: cfg.ReplicationPassword,
		Flavor:   "mysql",
		ServerID: 100, // Should be unique
		Dump: canal.DumpConfig{
			ExecutionPath: "", // We don't want to dump, just sync binlog
		},
		IncludeTableRegex: tableRegex,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create canal: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	l := &BinlogListener{
		cfg:       cfg,
		canal:     c,
		eventChan: make(chan BinlogEvent, 10000), // Buffered queue 10K
		ctx:       ctx,
		cancel:    cancel,
		tables:    tableMap,
	}

	c.SetEventHandler(&eventHandler{listener: l})

	return l, nil
}

func (l *BinlogListener) Start() error {
	logger.Log.Info("Starting binlog listener", zap.String("host", l.cfg.Host))
	
	// Start canal in a goroutine
	go func() {
		if err := l.canal.Run(); err != nil {
			logger.Log.Error("Canal run error", zap.Error(err))
		}
	}()
	
	return nil
}

func (l *BinlogListener) Stop() {
	l.cancel()
	l.canal.Close()
	close(l.eventChan)
	logger.Log.Info("Stopped binlog listener")
}

func (l *BinlogListener) Events() <-chan BinlogEvent {
	return l.eventChan
}

type eventHandler struct {
	canal.DummyEventHandler
	listener *BinlogListener
}

func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	// Filter tables if needed (though regex should handle it)
	if _, ok := h.listener.tables[e.Table.Name]; !ok {
		return nil
	}

	var eventType EventType
	switch e.Action {
	case canal.InsertAction:
		eventType = Insert
	case canal.UpdateAction:
		eventType = Update
	case canal.DeleteAction:
		eventType = Delete
	default:
		return nil
	}

	// Get current binlog position
	pos := h.listener.canal.SyncedPosition()

	binlogEvent := BinlogEvent{
		Type:       eventType,
		Schema:     e.Table.Schema,
		Table:      e.Table.Name,
		Rows:       e.Rows,
		Timestamp:  e.Header.Timestamp,
		BinlogFile: pos.Name,
		BinlogPos:  pos.Pos,
	}

	// Non-blocking send or block? Spec says "Push changes to buffered queue"
	// If queue is full, we should probably block to apply backpressure
	select {
	case h.listener.eventChan <- binlogEvent:
	case <-h.listener.ctx.Done():
		return h.listener.ctx.Err()
	}

	return nil
}

func (h *eventHandler) String() string {
	return "BinlogEventHandler"
}
