package config

import (
	"time"
)

type Config struct {
	Databases    DatabasesConfig `mapstructure:"databases"`
	StateStorage StateStorage    `mapstructure:"state_storage"`
	Sync         SyncConfig      `mapstructure:"sync"`
	Scheduler    SchedulerConfig `mapstructure:"scheduler"`
	Server       ServerConfig    `mapstructure:"server"`
	Logging      LoggingConfig   `mapstructure:"logging"`
}

type DatabasesConfig struct {
	Local DatabaseConnection `mapstructure:"local"`
	Cloud DatabaseConnection `mapstructure:"cloud"`
}

type DatabaseConnection struct {
	Host                string `mapstructure:"host"`
	Port                int    `mapstructure:"port"`
	User                string `mapstructure:"user"`
	Password            string `mapstructure:"password"`
	Database            string `mapstructure:"database"`
	ReplicationUser     string `mapstructure:"replication_user"`
	ReplicationPassword string `mapstructure:"replication_password"`
}

type StateStorage struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	FilePath string `mapstructure:"file_path"` // For SQLite
}

type SyncConfig struct {
	Mode            string        `mapstructure:"mode"`
	Tables          []TableConfig `mapstructure:"tables"`
	Workers         int           `mapstructure:"workers"`
	Realtime        bool          `mapstructure:"realtime"`
	BatchInsertSize int           `mapstructure:"batch_insert_size"`
}

type TableConfig struct {
	Name               string `mapstructure:"name"`
	ConflictResolution string `mapstructure:"conflict_resolution"`
	BatchSize          int    `mapstructure:"batch_size"`
	PrimaryKey         string `mapstructure:"primary_key"`
	TimestampColumn    string `mapstructure:"timestamp_column"`
}

type SchedulerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Interval string `mapstructure:"interval"`
}

type ServerConfig struct {
	Port         int      `mapstructure:"port"`
	Host         string   `mapstructure:"host"`
	AuthToken    string   `mapstructure:"auth_token"`
	ReadTimeout  string   `mapstructure:"read_timeout"`
	WriteTimeout string   `mapstructure:"write_timeout"`
	CorsOrigins  []string `mapstructure:"cors_origins"`
}

func (s ServerConfig) GetReadTimeout() time.Duration {
	d, _ := time.ParseDuration(s.ReadTimeout)
	return d
}

func (s ServerConfig) GetWriteTimeout() time.Duration {
	d, _ := time.ParseDuration(s.WriteTimeout)
	return d
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}
