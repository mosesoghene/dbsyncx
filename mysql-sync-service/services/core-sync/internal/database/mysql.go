package database

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

type Database struct {
	DB     *sql.DB
	Config config.DatabaseConnection
}

func NewDatabase(cfg config.DatabaseConnection) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	logger.Log.Info("Connected to database",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Database),
	)

	return &Database{
		DB:     db,
		Config: cfg,
	}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}

// ExecTx executes a function within a transaction
func (d *Database) ExecTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
