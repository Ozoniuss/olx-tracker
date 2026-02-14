package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type BaseRepositoryTestSuite struct {
	suite.Suite
	DB *sql.DB
}

func TestBaseRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(BaseRepositoryTestSuite))
}

func (s *BaseRepositoryTestSuite) TestNewUserAndGetUserID() {
	ctx := context.Background()
	username := "test-user"
	password := "test-password"

	createdUserID, err := NewUser(ctx, s.DB, username, password, false)
	s.Require().NoError(err)
	s.Require().NotEqual(uuid.Nil, createdUserID)
	s.T().Logf("got user UUID: %v", createdUserID)

	retrievedUserID, err := GetUserID(ctx, s.DB, username, password)
	s.Require().NoError(err)
	s.Equal(createdUserID, retrievedUserID)

	_, err = NewUser(ctx, s.DB, username, password, false)
	s.Require().Error(err)
	s.ErrorIs(err, ErrAlreadyExists)

	_, err = GetUserID(ctx, s.DB, "invalid-user", "invalid-password")
	s.Require().Error(err)
	s.ErrorIs(err, ErrNotFound)
}

func (s *BaseRepositoryTestSuite) SetupSuite() {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		s.T().Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run them.")
	}

	dbURL := testDatabaseURL()

	db, err := sql.Open("postgres", dbURL)
	s.Require().NoError(err)

	err = db.Ping()
	s.Require().NoError(err)
	s.DB = db
}

func (s *BaseRepositoryTestSuite) AfterTest(suiteName, testName string) {
	_, err := s.DB.Exec("DELETE FROM product_versions")
	s.Require().NoError(err)

	_, err = s.DB.Exec("DELETE FROM products")
	s.Require().NoError(err)

	_, err = s.DB.Exec("DELETE FROM users")
	s.Require().NoError(err)
}

func (s *BaseRepositoryTestSuite) TearDownSuite() {
	if s.DB != nil {
		s.DB.Close()
	}
}

func testDatabaseURL() string {
	if databaseURL := os.Getenv("TEST_DATABASE_URL"); databaseURL != "" {
		return databaseURL
	}

	user := getenvOrDefault("OLXTRACKER_POSTGRES_USER", "olxtracker")
	password := getenvOrDefault("OLXTRACKER_POSTGRES_PASSWORD", "olxtracker")
	host := getenvOrDefault("OLXTRACKER_POSTGRES_HOST", "127.0.0.1")
	port := getenvOrDefault("OLXTRACKER_POSTGRES_PORT", "5433")
	database := getenvOrDefault("OLXTRACKER_POSTGRES_DATABASE", "olxtracker")
	schema := getenvOrDefault("OLXTRACKER_POSTGRES_SCHEMA", "olxtracker")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		user,
		password,
		host,
		port,
		database,
		schema,
	)
}

func getenvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
