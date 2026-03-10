// Package database provides SQLite database operations for credential and data storage.
package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kertasbaru/me-cli-sunset/internal/model"
	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB connection and provides application-specific methods.
type DB struct {
	conn *sql.DB
}

// New creates a new database connection and runs migrations.
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate creates tables if they do not exist.
func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			number INTEGER UNIQUE NOT NULL,
			subscriber_id TEXT NOT NULL DEFAULT '',
			subscription_type TEXT NOT NULL DEFAULT '',
			refresh_token TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS bookmarks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			family_code TEXT NOT NULL,
			family_name TEXT NOT NULL DEFAULT '',
			is_enterprise INTEGER NOT NULL DEFAULT 0,
			variant_name TEXT NOT NULL,
			option_name TEXT NOT NULL DEFAULT '',
			"order" INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			UNIQUE(account_id, family_code, variant_name, "order")
		)`,
		`CREATE TABLE IF NOT EXISTS device_fingerprints (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fingerprint TEXT NOT NULL,
			device_id TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sentry_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			quotas TEXT NOT NULL,
			logged_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
		)`,
	}

	for _, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

// --- Account Operations ---

// GetAllAccounts returns all stored accounts.
func (db *DB) GetAllAccounts() ([]model.Account, error) {
	rows, err := db.conn.Query(
		`SELECT id, number, subscriber_id, subscription_type, refresh_token, is_active, created_at, updated_at
		FROM accounts ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var a model.Account
		if err := rows.Scan(&a.ID, &a.Number, &a.SubscriberID, &a.SubscriptionType, &a.RefreshToken, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// GetAccountByNumber returns an account by phone number.
func (db *DB) GetAccountByNumber(number int64) (*model.Account, error) {
	var a model.Account
	err := db.conn.QueryRow(
		`SELECT id, number, subscriber_id, subscription_type, refresh_token, is_active, created_at, updated_at
		FROM accounts WHERE number = ?`, number,
	).Scan(&a.ID, &a.Number, &a.SubscriberID, &a.SubscriptionType, &a.RefreshToken, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// GetActiveAccount returns the currently active account.
func (db *DB) GetActiveAccount() (*model.Account, error) {
	var a model.Account
	err := db.conn.QueryRow(
		`SELECT id, number, subscriber_id, subscription_type, refresh_token, is_active, created_at, updated_at
		FROM accounts WHERE is_active = 1 LIMIT 1`,
	).Scan(&a.ID, &a.Number, &a.SubscriberID, &a.SubscriptionType, &a.RefreshToken, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// UpsertAccount inserts or updates an account.
func (db *DB) UpsertAccount(number int64, subscriberID, subscriptionType, refreshToken string) error {
	now := time.Now()
	_, err := db.conn.Exec(
		`INSERT INTO accounts (number, subscriber_id, subscription_type, refresh_token, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(number) DO UPDATE SET
			subscriber_id = excluded.subscriber_id,
			subscription_type = excluded.subscription_type,
			refresh_token = excluded.refresh_token,
			updated_at = ?`,
		number, subscriberID, subscriptionType, refreshToken, now, now,
	)
	return err
}

// SetActiveAccount sets the specified account as active and deactivates all others.
func (db *DB) SetActiveAccount(number int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE accounts SET is_active = 0`); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE accounts SET is_active = 1 WHERE number = ?`, number); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteAccount removes an account by phone number.
func (db *DB) DeleteAccount(number int64) error {
	_, err := db.conn.Exec(`DELETE FROM accounts WHERE number = ?`, number)
	return err
}

// --- Bookmark Operations ---

// GetBookmarks returns all bookmarks for an account.
func (db *DB) GetBookmarks(accountID int64) ([]model.Bookmark, error) {
	rows, err := db.conn.Query(
		`SELECT id, account_id, family_code, family_name, is_enterprise, variant_name, option_name, "order", created_at
		FROM bookmarks WHERE account_id = ? ORDER BY id`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []model.Bookmark
	for rows.Next() {
		var b model.Bookmark
		if err := rows.Scan(&b.ID, &b.AccountID, &b.FamilyCode, &b.FamilyName, &b.IsEnterprise, &b.VariantName, &b.OptionName, &b.Order, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, rows.Err()
}

// AddBookmark adds a new bookmark for an account.
func (db *DB) AddBookmark(accountID int64, familyCode, familyName string, isEnterprise bool, variantName, optionName string, order int) error {
	_, err := db.conn.Exec(
		`INSERT OR IGNORE INTO bookmarks (account_id, family_code, family_name, is_enterprise, variant_name, option_name, "order")
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		accountID, familyCode, familyName, isEnterprise, variantName, optionName, order,
	)
	return err
}

// DeleteBookmark removes a bookmark by ID.
func (db *DB) DeleteBookmark(bookmarkID int64) error {
	_, err := db.conn.Exec(`DELETE FROM bookmarks WHERE id = ?`, bookmarkID)
	return err
}

// --- Device Fingerprint Operations ---

// GetFingerprint returns the stored device fingerprint.
func (db *DB) GetFingerprint() (*model.DeviceFingerprint, error) {
	var fp model.DeviceFingerprint
	err := db.conn.QueryRow(
		`SELECT id, fingerprint, device_id, created_at FROM device_fingerprints ORDER BY id DESC LIMIT 1`,
	).Scan(&fp.ID, &fp.Fingerprint, &fp.DeviceID, &fp.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fp, nil
}

// SaveFingerprint stores a new device fingerprint.
func (db *DB) SaveFingerprint(fingerprint, deviceID string) error {
	_, err := db.conn.Exec(
		`INSERT INTO device_fingerprints (fingerprint, device_id) VALUES (?, ?)`,
		fingerprint, deviceID,
	)
	return err
}

// --- Sentry Log Operations ---

// AddSentryLog stores a quota monitoring log entry.
func (db *DB) AddSentryLog(accountID int64, quotasJSON string) error {
	_, err := db.conn.Exec(
		`INSERT INTO sentry_logs (account_id, quotas) VALUES (?, ?)`,
		accountID, quotasJSON,
	)
	return err
}

// GetSentryLogs returns sentry logs for an account within a time range.
func (db *DB) GetSentryLogs(accountID int64, since, until time.Time) ([]map[string]interface{}, error) {
	rows, err := db.conn.Query(
		`SELECT quotas, logged_at FROM sentry_logs WHERE account_id = ? AND logged_at BETWEEN ? AND ? ORDER BY logged_at`,
		accountID, since, until,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var quotas string
		var loggedAt time.Time
		if err := rows.Scan(&quotas, &loggedAt); err != nil {
			return nil, err
		}
		logs = append(logs, map[string]interface{}{
			"quotas":    quotas,
			"logged_at": loggedAt,
		})
	}
	return logs, rows.Err()
}
