package mysqlhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	tests := []struct {
		name          string
		dsn           string
		wantErr       bool
		errorContains string
	}{
		{
			name:          "empty DSN",
			dsn:           "",
			wantErr:       true,
			errorContains: "failed to connect to MySQL",
		},
		{
			name:          "invalid DSN format",
			dsn:           "invalid-dsn",
			wantErr:       true,
			errorContains: "failed to connect to MySQL",
		},
		{
			name:          "malformed DSN",
			dsn:           "user:pass@/dbname",
			wantErr:       true,
			errorContains: "failed to connect to MySQL",
		},
		{
			name:          "non-existent host",
			dsn:           "user:pass@tcp(nonexistent.host:3306)/dbname",
			wantErr:       true,
			errorContains: "failed to connect to MySQL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := Connect(tt.dsn)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				if db != nil {
					db.Close()
				}
			}
		})
	}
}

func TestCheckConnection(t *testing.T) {
	tests := []struct {
		name          string
		dsn           string
		wantErr       bool
		errorContains string
	}{
		{
			name:          "empty DSN",
			dsn:           "",
			wantErr:       true,
			errorContains: "failed to connect to MySQL",
		},
		{
			name:          "invalid DSN format",
			dsn:           "invalid-dsn",
			wantErr:       true,
			errorContains: "failed to connect to MySQL",
		},
		{
			name:          "malformed DSN",
			dsn:           "user:pass@/dbname",
			wantErr:       true,
			errorContains: "failed to connect to MySQL",
		},
		{
			name:          "non-existent host",
			dsn:           "user:pass@tcp(nonexistent.host:3306)/dbname",
			wantErr:       true,
			errorContains: "failed to connect to MySQL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckConnection(tt.dsn)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCheckConnection_Integration provides an integration test that would work with a real MySQL instance
// This test is skipped by default and can be enabled by setting the INTEGRATION_TEST environment variable
func TestCheckConnection_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a real MySQL connection, but we can demonstrate the concept
	// In a real scenario, you'd set up a test database or use testcontainers
	t.Skip("Integration test requires real MySQL instance - use testcontainers or docker-compose for full integration testing")
	
	// Example of how this would work:
	// validDSN := os.Getenv("TEST_MYSQL_DSN")
	// if validDSN == "" {
	// 	t.Skip("TEST_MYSQL_DSN not set, skipping MySQL integration test")
	// }
	// 
	// err := CheckConnection(validDSN)
	// assert.NoError(t, err)
}