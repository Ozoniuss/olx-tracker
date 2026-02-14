package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Ozoniuss/olx-tracker/config"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAlreadyExists = fmt.Errorf("already exists")
	ErrNotFound      = fmt.Errorf("not found")
)

func GetPostgresURL(cfg config.PostgresConfig) string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.Schema)
	return dsn
}

func ConnectToPostgres(ctx context.Context, dsn string) (*sql.DB, error) {
	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test database connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

type ProductWithUrl struct {
	ID  uuid.UUID
	URL string
}

type ProductSnapshot struct {
	ID             uuid.UUID
	ProductID      uuid.UUID
	Version        int
	RetrievedAt    time.Time
	Name           string
	Description    string
	PriceSmallUnit int64
	Currency       string
	Availability   string
	RawJSON        []byte
}

func ListTrackedProductsForUser(ctx context.Context, db *sql.DB, userID uuid.UUID) ([]ProductWithUrl, error) {
	const query = `
		SELECT id, url
		FROM products
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var tracked []ProductWithUrl
	for rows.Next() {
		var p ProductWithUrl
		if err := rows.Scan(&p.ID, &p.URL); err != nil {
			return nil, err
		}
		tracked = append(tracked, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return tracked, nil
}

func TrackAddForUser(ctx context.Context, db *sql.DB, userID uuid.UUID, url string) error {
	const query = `
		INSERT INTO products (id, user_id, url)
		VALUES ($1, $2, $3)
	`

	_, err := db.ExecContext(ctx, query, uuid.New(), userID, url)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

func StoreNextAddSnapshot(
	ctx context.Context,
	db *sql.DB,
	productID uuid.UUID,
	name string,
	description string,
	priceSmallUnit int64,
	currency string,
	availability string,
	rawJSON []byte,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const selectQuery = `
		SELECT COALESCE(MAX(version), 0)
		FROM product_versions
		WHERE product_id = $1
	`

	var currentVersion int
	if err = tx.QueryRowContext(ctx, selectQuery, productID).Scan(&currentVersion); err != nil {
		return err
	}

	nextVersion := currentVersion + 1 // first version will be 1

	const insertQuery = `
		INSERT INTO product_versions (
			id,
			product_id,
			version,
			name,
			description,
			price_small_unit,
			currency,
			availability,
			raw_json
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = tx.ExecContext(
		ctx,
		insertQuery,
		uuid.New(),
		productID,
		nextVersion,
		name,
		description,
		priceSmallUnit,
		currency,
		availability,
		rawJSON,
	)
	if err != nil {
		var pgErr *pq.Error
		// optimistic locking concurrency control
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return err
	}

	return tx.Commit()
}

func ListAddSnapshotsForUser(
	ctx context.Context,
	db *sql.DB,
	userID uuid.UUID,
	productID uuid.UUID,
) ([]ProductSnapshot, error) {
	const query = `
		SELECT
			pv.id,
			pv.product_id,
			pv.version,
			pv.retrieved_at,
			pv.name,
			pv.description,
			pv.price_small_unit,
			pv.currency,
			pv.availability,
			pv.raw_json
		FROM products p
		INNER JOIN product_versions pv
			ON pv.product_id = p.id
		WHERE p.user_id = $1 AND p.id = $2
		ORDER BY pv.version DESC
	`

	rows, err := db.QueryContext(ctx, query, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var snapshots []ProductSnapshot
	for rows.Next() {
		var snapshot ProductSnapshot
		if err := rows.Scan(
			&snapshot.ID,
			&snapshot.ProductID,
			&snapshot.Version,
			&snapshot.RetrievedAt,
			&snapshot.Name,
			&snapshot.Description,
			&snapshot.PriceSmallUnit,
			&snapshot.Currency,
			&snapshot.Availability,
			&snapshot.RawJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		snapshots = append(snapshots, snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return snapshots, nil
}

func GetUserID(ctx context.Context, db *sql.DB, username, password string) (uuid.UUID, error) {
	const query = `
		SELECT id
		FROM users
		WHERE username = $1 AND password_hash = $2
	`

	var userID uuid.UUID
	err := db.QueryRowContext(ctx, query, username, password).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return userID, nil
}

func NewUser(ctx context.Context, db *sql.DB, username, password string, shouldHash bool) (uuid.UUID, error) {

	var passwordHash string
	if shouldHash {
		passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = string(passwordHashBytes)
	} else {
		passwordHash = password
	}

	const query = `
		INSERT INTO users (id, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var userID uuid.UUID
	err := db.QueryRowContext(ctx, query, uuid.New(), username, passwordHash).Scan(&userID)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, ErrAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return userID, nil
}
