package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
	cfg := &Config{
		Environment:             "development",
		Port:                    "8080",
		LogLevel:                "info",
		DatabaseHost:            "localhost",
		DatabasePort:            "5432",
		DatabaseUser:            "test",
		DatabasePassword:        "test",
		DatabaseName:            "test",
		DatabaseSSLMode:         "disable",
		DBMaxIdleConns:          10,
		DBMaxOpenConns:          100,
		DBConnMaxLifetime:       "1h",
		DBConnMaxIdleTime:       "30m",
		JWTSecret:               "test-secret-key-for-testing-only-32chars",
		JWTAccessTokenDuration:  "15m",
		JWTRefreshTokenDuration: "7d",
		JWTIssuer:               "test",
		EmailProvider:           "smtp",
		EmailFrom:               "test@example.com",
		EmailFromName:           "Test App",
		SMTPHost:                "localhost",
		SMTPPort:                587,
		FrontendURL:             "http://localhost:3000",
		BackendURL:              "http://localhost:8080",
		CSRFSecret:              "test-csrf-secret-32-characters-long",
		StorageProvider:         "local",
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfigHelperMethods(t *testing.T) {
	cfg := &Config{
		Environment:             "development",
		JWTAccessTokenDuration:  "15m",
		JWTRefreshTokenDuration: "7d",
		DBConnMaxLifetime:       "1h",
		DBConnMaxIdleTime:       "30m",
		DatabaseHost:            "localhost",
		DatabasePort:            "5432",
		DatabaseUser:            "test",
		DatabasePassword:        "test",
		DatabaseName:            "test_db",
		DatabaseSSLMode:         "disable",
	}

	assert.True(t, cfg.IsDevelopment())
	assert.False(t, cfg.IsProduction())
	assert.Contains(t, cfg.DatabaseDSN(), "test_db")
	assert.Equal(t, "15m0s", cfg.JWTAccessTokenDurationParsed().String())
}

func TestDatabaseDSN(t *testing.T) {
	cfg := &Config{
		DatabaseHost:     "localhost",
		DatabasePort:     "5432",
		DatabaseUser:     "testuser",
		DatabasePassword: "testpass",
		DatabaseName:     "testdb",
		DatabaseSSLMode:  "disable",
	}

	dsn := cfg.DatabaseDSN()
	assert.Contains(t, dsn, "localhost")
	assert.Contains(t, dsn, "5432")
	assert.Contains(t, dsn, "testdb")
	assert.Contains(t, dsn, "sslmode=disable")
}
