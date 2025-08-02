package seed

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Seed represents a database seed
type Seed struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"uniqueIndex;not null"`
	Applied   bool   `gorm:"not null;default:false"`
	AppliedAt *time.Time
	Checksum  string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SeederFunc represents a function that performs seeding
type SeederFunc func(ctx context.Context, db *gorm.DB) error

// SeederDefinition defines a seeder with its function
type SeederDefinition struct {
	Name        string
	Description string
	Fn          SeederFunc
	// Dependencies list other seeders this one depends on
	Dependencies []string
	// Environment specifies which environments this seeder should run in
	// If empty, runs in all environments
	Environment []string
}

// Seeder handles database seeding
type Seeder struct {
	db      *gorm.DB
	seeders map[string]SeederDefinition
	env     string
}

// NewSeeder creates a new seeder instance
func NewSeeder(db *gorm.DB, environment string) *Seeder {
	return &Seeder{
		db:      db,
		seeders: make(map[string]SeederDefinition),
		env:     environment,
	}
}

// AddSeeder adds a seeder to the seeder registry
func (s *Seeder) AddSeeder(seeder SeederDefinition) {
	s.seeders[seeder.Name] = seeder
}

// SeedAll runs all applicable seeders for the current environment
func (s *Seeder) SeedAll(ctx context.Context) error {
	// Ensure seeds table exists
	if err := s.db.AutoMigrate(&Seed{}); err != nil {
		return err
	}

	// Get all applicable seeders for current environment
	applicable := s.getApplicableSeeders()

	// Resolve dependencies and get execution order
	ordered, err := s.resolveDependencies(applicable)
	if err != nil {
		return err
	}

	// Get already applied seeds
	applied, err := s.getAppliedSeeds(ctx)
	if err != nil {
		return err
	}

	// Run pending seeders
	for _, seederName := range ordered {
		if applied[seederName] {
			continue
		}

		seeder := s.seeders[seederName]
		if err := s.runSeeder(ctx, seeder); err != nil {
			return err
		}
	}

	return nil
}

// SeedSpecific runs a specific seeder by name
func (s *Seeder) SeedSpecific(ctx context.Context, name string) error {
	seeder, exists := s.seeders[name]
	if !exists {
		return ErrSeederNotFound
	}

	// Check if already applied
	applied, err := s.getAppliedSeeds(ctx)
	if err != nil {
		return err
	}

	if applied[name] {
		return ErrSeederAlreadyApplied
	}

	return s.runSeeder(ctx, seeder)
}

// getApplicableSeeders returns seeders that should run in current environment
func (s *Seeder) getApplicableSeeders() []string {
	var applicable []string
	for name, seeder := range s.seeders {
		if len(seeder.Environment) == 0 {
			// Run in all environments
			applicable = append(applicable, name)
			continue
		}

		for _, env := range seeder.Environment {
			if env == s.env {
				applicable = append(applicable, name)
				break
			}
		}
	}
	return applicable
}

// resolveDependencies resolves seeder dependencies and returns execution order
func (s *Seeder) resolveDependencies(seeders []string) ([]string, error) {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []string

	var visit func(string) error
	visit = func(name string) error {
		if visiting[name] {
			return ErrCircularDependency
		}
		if visited[name] {
			return nil
		}

		visiting[name] = true

		seeder, exists := s.seeders[name]
		if !exists {
			return ErrSeederNotFound
		}

		for _, dep := range seeder.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true
		result = append(result, name)

		return nil
	}

	for _, name := range seeders {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// getAppliedSeeds returns a map of applied seed names
func (s *Seeder) getAppliedSeeds(ctx context.Context) (map[string]bool, error) {
	var seeds []Seed
	if err := s.db.WithContext(ctx).Where("applied = ?", true).Find(&seeds).Error; err != nil {
		return nil, err
	}

	applied := make(map[string]bool)
	for _, seed := range seeds {
		applied[seed.Name] = true
	}

	return applied, nil
}

// runSeeder executes a seeder and records it as applied
func (s *Seeder) runSeeder(ctx context.Context, seeder SeederDefinition) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Run the seeder
		if err := seeder.Fn(ctx, tx); err != nil {
			return err
		}

		// Record as applied
		now := time.Now()
		seedRecord := Seed{
			Name:      seeder.Name,
			Applied:   true,
			AppliedAt: &now,
			Checksum:  generateChecksum(seeder.Name + seeder.Description),
		}

		return tx.WithContext(ctx).Create(&seedRecord).Error
	})
}

// generateChecksum generates a simple checksum for seed tracking
func generateChecksum(data string) string {
	hash := uint32(0)
	for _, c := range data {
		hash = hash*31 + uint32(c)
	}
	return string(rune(hash))
}
