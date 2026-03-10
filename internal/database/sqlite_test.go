package database

import (
	"os"
	"testing"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(dbPath)
	})
	return db
}

func TestAccountOperations(t *testing.T) {
	db := setupTestDB(t)

	// Test empty accounts
	accounts, err := db.GetAllAccounts()
	if err != nil {
		t.Fatalf("GetAllAccounts failed: %v", err)
	}
	if len(accounts) != 0 {
		t.Fatalf("Expected 0 accounts, got %d", len(accounts))
	}

	// Test upsert account
	err = db.UpsertAccount(6281234567890, "sub1", "PREPAID", "refresh_token_1")
	if err != nil {
		t.Fatalf("UpsertAccount failed: %v", err)
	}

	// Test get by number
	account, err := db.GetAccountByNumber(6281234567890)
	if err != nil {
		t.Fatalf("GetAccountByNumber failed: %v", err)
	}
	if account == nil {
		t.Fatal("Expected account, got nil")
	}
	if account.Number != 6281234567890 {
		t.Fatalf("Expected number 6281234567890, got %d", account.Number)
	}
	if account.SubscriberID != "sub1" {
		t.Fatalf("Expected subscriber_id sub1, got %s", account.SubscriberID)
	}
	if account.SubscriptionType != "PREPAID" {
		t.Fatalf("Expected PREPAID, got %s", account.SubscriptionType)
	}

	// Test upsert (update existing)
	err = db.UpsertAccount(6281234567890, "sub1_updated", "PRIORITAS", "refresh_token_2")
	if err != nil {
		t.Fatalf("UpsertAccount (update) failed: %v", err)
	}

	account, err = db.GetAccountByNumber(6281234567890)
	if err != nil {
		t.Fatalf("GetAccountByNumber after update failed: %v", err)
	}
	if account.RefreshToken != "refresh_token_2" {
		t.Fatalf("Expected updated refresh token, got %s", account.RefreshToken)
	}
	if account.SubscriptionType != "PRIORITAS" {
		t.Fatalf("Expected PRIORITAS, got %s", account.SubscriptionType)
	}

	// Test set active
	err = db.UpsertAccount(6281234567891, "sub2", "PREPAID", "refresh_token_3")
	if err != nil {
		t.Fatalf("UpsertAccount (second) failed: %v", err)
	}

	err = db.SetActiveAccount(6281234567891)
	if err != nil {
		t.Fatalf("SetActiveAccount failed: %v", err)
	}

	active, err := db.GetActiveAccount()
	if err != nil {
		t.Fatalf("GetActiveAccount failed: %v", err)
	}
	if active == nil {
		t.Fatal("Expected active account, got nil")
	}
	if active.Number != 6281234567891 {
		t.Fatalf("Expected active number 6281234567891, got %d", active.Number)
	}

	// Test delete
	err = db.DeleteAccount(6281234567890)
	if err != nil {
		t.Fatalf("DeleteAccount failed: %v", err)
	}

	accounts, err = db.GetAllAccounts()
	if err != nil {
		t.Fatalf("GetAllAccounts after delete failed: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("Expected 1 account after delete, got %d", len(accounts))
	}

	// Test get non-existent
	account, err = db.GetAccountByNumber(6281234567890)
	if err != nil {
		t.Fatalf("GetAccountByNumber (non-existent) failed: %v", err)
	}
	if account != nil {
		t.Fatal("Expected nil for deleted account")
	}
}

func TestBookmarkOperations(t *testing.T) {
	db := setupTestDB(t)

	// Create an account first
	err := db.UpsertAccount(6281234567890, "sub1", "PREPAID", "rt1")
	if err != nil {
		t.Fatalf("UpsertAccount failed: %v", err)
	}

	account, _ := db.GetAccountByNumber(6281234567890)

	// Test empty bookmarks
	bookmarks, err := db.GetBookmarks(account.ID)
	if err != nil {
		t.Fatalf("GetBookmarks failed: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Fatalf("Expected 0 bookmarks, got %d", len(bookmarks))
	}

	// Test add bookmark
	err = db.AddBookmark(account.ID, "FAM001", "Test Family", false, "Variant A", "Option 1", 1)
	if err != nil {
		t.Fatalf("AddBookmark failed: %v", err)
	}

	bookmarks, err = db.GetBookmarks(account.ID)
	if err != nil {
		t.Fatalf("GetBookmarks after add failed: %v", err)
	}
	if len(bookmarks) != 1 {
		t.Fatalf("Expected 1 bookmark, got %d", len(bookmarks))
	}
	if bookmarks[0].FamilyCode != "FAM001" {
		t.Fatalf("Expected family code FAM001, got %s", bookmarks[0].FamilyCode)
	}

	// Test duplicate bookmark (should be ignored)
	err = db.AddBookmark(account.ID, "FAM001", "Test Family", false, "Variant A", "Option 1", 1)
	if err != nil {
		t.Fatalf("AddBookmark (duplicate) failed: %v", err)
	}

	bookmarks, err = db.GetBookmarks(account.ID)
	if err != nil {
		t.Fatalf("GetBookmarks after duplicate failed: %v", err)
	}
	if len(bookmarks) != 1 {
		t.Fatalf("Expected 1 bookmark (duplicate ignored), got %d", len(bookmarks))
	}

	// Test delete bookmark
	err = db.DeleteBookmark(bookmarks[0].ID)
	if err != nil {
		t.Fatalf("DeleteBookmark failed: %v", err)
	}

	bookmarks, err = db.GetBookmarks(account.ID)
	if err != nil {
		t.Fatalf("GetBookmarks after delete failed: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Fatalf("Expected 0 bookmarks after delete, got %d", len(bookmarks))
	}
}

func TestFingerprintOperations(t *testing.T) {
	db := setupTestDB(t)

	// Test empty fingerprint
	fp, err := db.GetFingerprint()
	if err != nil {
		t.Fatalf("GetFingerprint failed: %v", err)
	}
	if fp != nil {
		t.Fatal("Expected nil fingerprint")
	}

	// Test save fingerprint
	err = db.SaveFingerprint("test_fingerprint_value", "test_device_id")
	if err != nil {
		t.Fatalf("SaveFingerprint failed: %v", err)
	}

	fp, err = db.GetFingerprint()
	if err != nil {
		t.Fatalf("GetFingerprint after save failed: %v", err)
	}
	if fp == nil {
		t.Fatal("Expected fingerprint, got nil")
	}
	if fp.Fingerprint != "test_fingerprint_value" {
		t.Fatalf("Expected fingerprint value, got %s", fp.Fingerprint)
	}
	if fp.DeviceID != "test_device_id" {
		t.Fatalf("Expected device ID, got %s", fp.DeviceID)
	}
}

func TestSentryLogOperations(t *testing.T) {
	db := setupTestDB(t)

	// Create an account first
	err := db.UpsertAccount(6281234567890, "sub1", "PREPAID", "rt1")
	if err != nil {
		t.Fatalf("UpsertAccount failed: %v", err)
	}
	account, _ := db.GetAccountByNumber(6281234567890)

	// Test add sentry log
	err = db.AddSentryLog(account.ID, `{"quota": "1GB"}`)
	if err != nil {
		t.Fatalf("AddSentryLog failed: %v", err)
	}
}
