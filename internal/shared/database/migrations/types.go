package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID          uint   `gorm:"primarykey"`
	Version     string `gorm:"uniqueIndex;not null"`
	Description string `gorm:"not null"`
	Applied     bool   `gorm:"not null;default:false"`
	AppliedAt   *time.Time
	Checksum    string `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MigrationFunc represents a function that performs a migration
type MigrationFunc func(ctx context.Context, db *gorm.DB) error

// MigrationDefinition defines a migration with its up and down functions
type MigrationDefinition struct {
	Version     string
	Description string
	Up          MigrationFunc
	Down        MigrationFunc
}

// Migrator handles database migrations
type Migrator struct {
	db         *gorm.DB
	migrations []MigrationDefinition
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]MigrationDefinition, 0),
	}
}

// AddMigration adds a migration to the migrator
func (m *Migrator) AddMigration(migration MigrationDefinition) {
	m.migrations = append(m.migrations, migration)
}

// GetPendingMigrations returns migrations that haven't been applied
func (m *Migrator) GetPendingMigrations(ctx context.Context) ([]MigrationDefinition, error) {
	// Ensure migrations table exists
	if err := m.db.AutoMigrate(&Migration{}); err != nil {
		return nil, err
	}

	var appliedMigrations []Migration
	if err := m.db.WithContext(ctx).Where("applied = ?", true).Find(&appliedMigrations).Error; err != nil {
		return nil, err
	}

	appliedVersions := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	var pending []MigrationDefinition
	for _, migration := range m.migrations {
		if !appliedVersions[migration.Version] {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// ApplyMigrations applies all pending migrations
func (m *Migrator) ApplyMigrations(ctx context.Context) error {
	pending, err := m.GetPendingMigrations(ctx)
	if err != nil {
		return err
	}

	for _, migration := range pending {
		if err := m.applyMigration(ctx, migration); err != nil {
			return err
		}
	}

	return nil
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(ctx context.Context, migration MigrationDefinition) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		// Run the migration
		if err := migration.Up(ctx, tx); err != nil {
			return err
		}

		// Record the migration as applied
		now := time.Now()
		migrationRecord := Migration{
			Version:     migration.Version,
			Description: migration.Description,
			Applied:     true,
			AppliedAt:   &now,
			Checksum:    generateChecksum(migration.Version + migration.Description),
		}

		return tx.WithContext(ctx).Create(&migrationRecord).Error
	})
}

// RollbackMigration rolls back the last applied migration
func (m *Migrator) RollbackMigration(ctx context.Context) error {
	var lastMigration Migration
	if err := m.db.WithContext(ctx).
		Where("applied = ?", true).
		Order("applied_at DESC").
		First(&lastMigration).Error; err != nil {
		return err
	}

	// Find the migration definition
	var migrationDef *MigrationDefinition
	for _, migration := range m.migrations {
		if migration.Version == lastMigration.Version {
			migrationDef = &migration
			break
		}
	}

	if migrationDef == nil {
		return ErrMigrationNotFound
	}

	return m.db.Transaction(func(tx *gorm.DB) error {
		// Run the rollback
		if err := migrationDef.Down(ctx, tx); err != nil {
			return err
		}

		// Mark migration as not applied
		return tx.WithContext(ctx).
			Model(&lastMigration).
			Updates(map[string]interface{}{
				"applied":    false,
				"applied_at": nil,
			}).Error
	})
}

// generateChecksum generates a simple checksum for migration tracking
func generateChecksum(data string) string {
	// Simple hash function for demonstration
	// In production, you might want to use a proper hash function
	hash := uint32(0)
	for _, c := range data {
		hash = hash*31 + uint32(c)
	}
	return string(rune(hash))
}
