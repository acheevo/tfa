package seed

import "errors"

var (
	// ErrSeederNotFound is returned when a seeder is not found
	ErrSeederNotFound = errors.New("seeder not found")

	// ErrSeederAlreadyApplied is returned when trying to apply an already applied seeder
	ErrSeederAlreadyApplied = errors.New("seeder already applied")

	// ErrCircularDependency is returned when there's a circular dependency in seeders
	ErrCircularDependency = errors.New("circular dependency detected in seeders")

	// ErrDependencyNotFound is returned when a dependency seeder is not found
	ErrDependencyNotFound = errors.New("dependency seeder not found")
)
