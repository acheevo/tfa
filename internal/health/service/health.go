package service

import (
	"log/slog"
	"time"

	"github.com/acheevo/tfa/internal/health/domain"
	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/database"
	"github.com/acheevo/tfa/internal/shared/health"
)

type HealthService struct {
	config *config.Config
	db     *database.DB
	logger *slog.Logger
}

func NewHealthService(config *config.Config, db *database.DB, logger *slog.Logger) *HealthService {
	return &HealthService{
		config: config,
		db:     db,
		logger: logger,
	}
}

func (s *HealthService) GetHealth() *domain.HealthStatus {
	services := make(map[string]interface{})

	dbStatus := string(health.StatusHealthy)
	if err := s.db.Ping(); err != nil {
		dbStatus = string(health.StatusUnhealthy)
		s.logger.Error("database health check failed", "error", err)
	}
	services["database"] = map[string]string{"status": dbStatus}

	overallStatus := string(health.StatusHealthy)
	if dbStatus != string(health.StatusHealthy) {
		overallStatus = string(health.StatusUnhealthy)
	}

	return &domain.HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
		Services:  services,
	}
}
