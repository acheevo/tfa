package service

import (
	"log/slog"
	"time"

	"github.com/acheevo/tfa/internal/info/domain"
	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/database"
)

type InfoService struct {
	config *config.Config
	db     *database.DB
	logger *slog.Logger
}

func NewInfoService(config *config.Config, db *database.DB, logger *slog.Logger) *InfoService {
	return &InfoService{
		config: config,
		db:     db,
		logger: logger,
	}
}

func (s *InfoService) GetInfo() *domain.Info {
	return &domain.Info{
		Name:        "Fullstack Template API",
		Version:     "1.0.0",
		Environment: s.config.Environment,
		BuildTime:   time.Now().UTC().Format(time.RFC3339),
	}
}
