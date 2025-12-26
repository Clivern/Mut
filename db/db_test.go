// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/clivern/mut/migration"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	testDBOnce sync.Once
	testDBConn *sql.DB
	testDBErr  error
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	testDBOnce.Do(func() {
		port := 5432
		if raw := os.Getenv("MUT_DATABASE_PORT"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				testDBErr = fmt.Errorf("invalid MUT_DATABASE_PORT: %w", err)
				return
			}

			port = parsed
		}

		config := DatabaseConfig{
			Driver:   envOr("MUT_DATABASE_DRIVER", "postgres"),
			Host:     envOr("MUT_DATABASE_HOST", "localhost"),
			Port:     port,
			Username: envOr("MUT_DATABASE_USERNAME", "postgres"),
			Password: envOr("MUT_DATABASE_PASSWORD", "postgres"),
			Database: envOr("MUT_DATABASE_NAME", "mut"),
		}

		var conn *Connection
		var err error
		for attempt := 1; attempt <= 5; attempt++ {
			conn, err = NewConnection(config)
			if err == nil {
				break
			}

			time.Sleep(500 * time.Millisecond)
		}
		if err != nil {
			testDBErr = err
			return
		}

		mgr := migration.NewManager(conn.DB)
		for _, item := range migration.GetAll() {
			mgr.Register(item)
		}
		if err := mgr.Up(); err != nil {
			_ = conn.Close()
			testDBErr = err
			return
		}

		testDBConn = conn.DB
	})

	if testDBErr != nil {
		t.Fatalf("postgres unavailable: %v", testDBErr)
	}

	return testDBConn
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func stringPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func createTestUser(t *testing.T, database *sql.DB) *User {
	t.Helper()

	user := &User{
		Name:        "tester",
		Email:       fmt.Sprintf("u-%s@t.test", uuid.NewString()[:8]),
		Password:    "hash",
		Role:        UserRoleRegular,
		IsActive:    true,
		Language:    UserLanguageEN,
		Theme:       UserThemeDefault,
		LastLoginAt: time.Now().UTC(),
	}
	require.NoError(t, NewUserRepository(database).Create(user))
	t.Cleanup(func() {
		_ = NewUserRepository(database).Delete(user.Id)
	})

	return user
}
