package migrations

import "errors"

var (
	// ErrMigrationNotFound is returned when a migration is not found
	ErrMigrationNotFound = errors.New("migration not found")

	// ErrMigrationAlreadyApplied is returned when trying to apply an already applied migration
	ErrMigrationAlreadyApplied = errors.New("migration already applied")

	// ErrNoMigrationsToRollback is returned when there are no migrations to rollback
	ErrNoMigrationsToRollback = errors.New("no migrations to rollback")

	// ErrInvalidMigrationVersion is returned when migration version is invalid
	ErrInvalidMigrationVersion = errors.New("invalid migration version")
)
