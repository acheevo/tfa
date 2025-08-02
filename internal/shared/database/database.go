package database

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/shared/database/migrations"
	"github.com/acheevo/tfa/internal/shared/database/seed"
	emaildomain "github.com/acheevo/tfa/internal/shared/email/domain"
)

type DB struct {
	*gorm.DB
	sqlDB    *sql.DB
	migrator *migrations.Migrator
	seeder   *seed.Seeder
	logger   *slog.Logger
}

func New(dsn string, isDevelopment bool, logger *slog.Logger, environment string) (*DB, error) {
	logLevel := gormlogger.Silent
	if isDevelopment {
		logLevel = gormlogger.Info
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	db := &DB{
		DB:       gormDB,
		sqlDB:    sqlDB,
		migrator: migrations.NewMigrator(gormDB),
		seeder:   seed.NewSeeder(gormDB, environment),
		logger:   logger,
	}

	// Initialize migrations
	db.initializeMigrations()

	// Initialize seeders
	db.initializeSeeders()

	// Auto-migrate authentication tables (legacy support)
	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) SetConnectionPool(maxIdleConns, maxOpenConns int, maxLifetime time.Duration) error {
	db.sqlDB.SetMaxIdleConns(maxIdleConns)
	db.sqlDB.SetMaxOpenConns(maxOpenConns)
	db.sqlDB.SetConnMaxLifetime(maxLifetime)
	return nil
}

func (db *DB) Close() error {
	return db.sqlDB.Close()
}

func (db *DB) Ping() error {
	return db.sqlDB.Ping()
}

// migrate runs database migrations for all models (legacy support)
func (db *DB) migrate() error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.RefreshToken{},
		&domain.PasswordReset{},
		&domain.AuditLog{},
		&emaildomain.QueuedEmail{},
		&emaildomain.EmailDeliveryEvent{},
	)
}

// GetMigrator returns the database migrator
func (db *DB) GetMigrator() *migrations.Migrator {
	return db.migrator
}

// GetSeeder returns the database seeder
func (db *DB) GetSeeder() *seed.Seeder {
	return db.seeder
}

// Migrate runs all pending migrations
func (db *DB) Migrate(ctx context.Context) error {
	db.logger.Info("Running database migrations...")
	if err := db.migrator.ApplyMigrations(ctx); err != nil {
		db.logger.Error("Failed to run migrations", "error", err)
		return err
	}
	db.logger.Info("Database migrations completed successfully")
	return nil
}

// Seed runs all applicable seeders
func (db *DB) Seed(ctx context.Context) error {
	db.logger.Info("Running database seeders...")
	if err := db.seeder.SeedAll(ctx); err != nil {
		db.logger.Error("Failed to run seeders", "error", err)
		return err
	}
	db.logger.Info("Database seeders completed successfully")
	return nil
}

// CheckHealth performs a comprehensive health check on the database
func (db *DB) CheckHealth(ctx context.Context) error {
	// Check basic connectivity
	if err := db.Ping(); err != nil {
		return err
	}

	// Check if we can perform a simple query
	var result int
	if err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return err
	}

	return nil
}

// GetConnectionStats returns database connection statistics
func (db *DB) GetConnectionStats() sql.DBStats {
	return db.sqlDB.Stats()
}

// initializeMigrations registers all application migrations
func (db *DB) initializeMigrations() {
	// Example migration - you can add more as needed
	db.migrator.AddMigration(migrations.MigrationDefinition{
		Version:     "20240101_000001",
		Description: "Create initial auth tables",
		Up: func(ctx context.Context, db *gorm.DB) error {
			// This migration is handled by AutoMigrate for now
			// You can add custom migration logic here
			return nil
		},
		Down: func(ctx context.Context, db *gorm.DB) error {
			// Rollback logic
			return nil
		},
	})
}

// initializeSeeders registers all application seeders
func (db *DB) initializeSeeders() {
	// Development user seeder
	db.seeder.AddSeeder(seed.SeederDefinition{
		Name:        "dev_users",
		Description: "Create development users",
		Environment: []string{"development"},
		Fn: func(ctx context.Context, db *gorm.DB) error {
			// Check if admin user already exists
			var count int64
			if err := db.Model(&domain.User{}).Where("email = ?", "admin@localhost").Count(&count).Error; err != nil {
				return err
			}

			if count > 0 {
				return nil // User already exists
			}

			// Create admin user
			admin := &domain.User{
				Email:         "admin@localhost",
				FirstName:     "Admin",
				LastName:      "User",
				EmailVerified: true,
				Role:          domain.RoleAdmin,
				PasswordHash:  "$2a$12$example_hashed_password", // bcrypt hash of "password"
			}

			return db.Create(admin).Error
		},
	})

	// Production setup seeder
	db.seeder.AddSeeder(seed.SeederDefinition{
		Name:        "production_setup",
		Description: "Setup production environment defaults",
		Environment: []string{"production"},
		Fn: func(ctx context.Context, db *gorm.DB) error {
			// Add any production-specific setup here
			return nil
		},
	})
}
