package sync

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	
	"mysql-sync-service/internal/config"
	"mysql-sync-service/internal/logger"
)

type Scheduler struct {
	cfg     config.SchedulerConfig
	manager *Manager
	cron    *cron.Cron
	entryID cron.EntryID
}

func NewScheduler(cfg config.SchedulerConfig, manager *Manager) *Scheduler {
	return &Scheduler{
		cfg:     cfg,
		manager: manager,
		cron:    cron.New(),
	}
}

func (s *Scheduler) Start() {
	if !s.cfg.Enabled {
		logger.Log.Info("Scheduler is disabled")
		return
	}
	
	logger.Log.Info("Starting scheduler", zap.String("interval", s.cfg.Interval))
	
	id, err := s.cron.AddFunc(s.cfg.Interval, func() {
		s.triggerSync()
	})
	
	if err != nil {
		logger.Log.Error("Failed to schedule job", zap.Error(err))
		return
	}
	
	s.entryID = id
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
	logger.Log.Info("Stopped scheduler")
}

func (s *Scheduler) triggerSync() {
	logger.Log.Info("Triggering scheduled sync")
	
	status := s.manager.GetStatus()
	if status == "running" {
		logger.Log.Info("Sync already running, skipping scheduled run")
		return
	}
	
	if err := s.manager.Start(); err != nil {
		logger.Log.Error("Failed to start scheduled sync", zap.Error(err))
	}
}
