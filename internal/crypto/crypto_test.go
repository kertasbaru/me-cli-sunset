package crypto

import (
	"strings"
	"testing"
	"time"
)

func newTestCryptoService() *CryptoService {
	return NewCryptoService(
		"1234567890123456",                 // xdataKey (16 bytes)
		"test_ax_api_sig_key",              // axAPISigKey
		"test_base_secret",                 // xAPIBaseSecret
		"1234567890123456",                 // encryptedFieldKey (16 bytes)
		"12345678901234567890123456789012", // axFPKey (32 bytes)
	)
}

func TestEncryptDecryptXData(t *testing.T) {
	cs := newTestCryptoService()
	plaintext := `{"test":"data","value":123}`
	xtimeMs := int64(1700000000000)

	encrypted, err := cs.EncryptXData(plaintext, xtimeMs)
	if err != nil {
		t.Fatalf("EncryptXData failed: %v", err)
	}

	if encrypted == "" {
		t.Fatal("Expected non-empty encrypted data")
	}

	decrypted, err := cs.DecryptXData(encrypted, xtimeMs)
	if err != nil {
		t.Fatalf("DecryptXData failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("Expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestMakeXSignature(t *testing.T) {
	cs := newTestCryptoService()

	sig := cs.MakeXSignature("test_id_token", "POST", "api/v8/profile", 1700000000)
	if sig == "" {
		t.Fatal("Expected non-empty signature")
	}

	// Signature should be hex-encoded SHA512 HMAC (128 chars)
	if len(sig) != 128 {
		t.Fatalf("Expected 128 char hex signature, got %d chars", len(sig))
	}
}

func TestMakeAXAPISignature(t *testing.T) {
	cs := newTestCryptoService()

	sig := cs.MakeAXAPISignature("2023-10-20T12:34:56.78+07:00", "6281234567890", "123456", "SMS")
	if sig == "" {
		t.Fatal("Expected non-empty signature")
	}
}

func TestBuildEncryptedField(t *testing.T) {
	cs := newTestCryptoService()

	field, err := cs.BuildEncryptedField(false)
	if err != nil {
		t.Fatalf("BuildEncryptedField failed: %v", err)
	}

	if field == "" {
		t.Fatal("Expected non-empty encrypted field")
	}

	// Should end with 16 hex chars (IV)
	if len(field) < 16 {
		t.Fatal("Encrypted field too short")
	}
}

func TestEncryptDecryptCircleMSISDN(t *testing.T) {
	cs := newTestCryptoService()
	msisdn := "6281234567890"

	encrypted, err := cs.EncryptCircleMSISDN(msisdn)
	if err != nil {
		t.Fatalf("EncryptCircleMSISDN failed: %v", err)
	}

	if encrypted == "" {
		t.Fatal("Expected non-empty encrypted MSISDN")
	}

	decrypted, err := cs.DecryptCircleMSISDN(encrypted)
	if err != nil {
		t.Fatalf("DecryptCircleMSISDN failed: %v", err)
	}

	if decrypted != msisdn {
		t.Fatalf("Expected '%s', got '%s'", msisdn, decrypted)
	}
}

func TestGenerateFingerprint(t *testing.T) {
	cs := newTestCryptoService()
	dev := DeviceInfo{
		Manufacturer:   "samsung",
		Model:          "SM-N935F",
		Lang:           "en",
		Resolution:     "720x1540",
		TZShort:        "GMT07:00",
		IP:             "192.169.69.69",
		FontScale:      1.0,
		AndroidRelease: "13",
		MSISDN:         "6281398370564",
	}

	fp, err := cs.GenerateFingerprint(dev)
	if err != nil {
		t.Fatalf("GenerateFingerprint failed: %v", err)
	}

	if fp == "" {
		t.Fatal("Expected non-empty fingerprint")
	}
}

func TestDeviceIDFromFingerprint(t *testing.T) {
	deviceID := DeviceIDFromFingerprint("test_fingerprint")
	if deviceID == "" {
		t.Fatal("Expected non-empty device ID")
	}
	// MD5 hex string = 32 chars
	if len(deviceID) != 32 {
		t.Fatalf("Expected 32 char hex device ID, got %d chars", len(deviceID))
	}
}

func TestJavaLikeTimestamp(t *testing.T) {
	gmt7 := time.FixedZone("GMT+7", 7*3600)
	testTime := time.Date(2023, 10, 20, 12, 34, 56, 0, gmt7)
	ts := JavaLikeTimestamp(testTime)
	if !strings.Contains(ts, "2023-10-20") {
		t.Fatalf("Unexpected timestamp format: %s", ts)
	}
	if !strings.Contains(ts, "12:34:56") {
		t.Fatalf("Unexpected timestamp format: %s", ts)
	}
}

func TestMakeXSignaturePayment(t *testing.T) {
	cs := newTestCryptoService()
	sig := cs.MakeXSignaturePayment("access_token", 1700000000, "PKG001", "PAY001", "BALANCE", "BUY_PACKAGE", "api/v8/settlements/buy")
	if sig == "" {
		t.Fatal("Expected non-empty payment signature")
	}
	if len(sig) != 128 {
		t.Fatalf("Expected 128 char hex signature, got %d chars", len(sig))
	}
}

func TestMakeXSignatureBounty(t *testing.T) {
	cs := newTestCryptoService()
	sig := cs.MakeXSignatureBounty("access_token", 1700000000, "PKG001", "PAY001")
	if sig == "" {
		t.Fatal("Expected non-empty bounty signature")
	}
}

func TestMakeXSignatureLoyalty(t *testing.T) {
	cs := newTestCryptoService()
	sig := cs.MakeXSignatureLoyalty(1700000000, "PKG001", "CONFIRM001", "api/v8/loyalty/redeem")
	if sig == "" {
		t.Fatal("Expected non-empty loyalty signature")
	}
}

func TestMakeXSignatureBountyAllotment(t *testing.T) {
	cs := newTestCryptoService()
	sig := cs.MakeXSignatureBountyAllotment(1700000000, "PKG001", "CONFIRM001", "api/v8/bounty/allot", "6281234567890")
	if sig == "" {
		t.Fatal("Expected non-empty bounty allotment signature")
	}
}

func TestMakeXSignatureBasic(t *testing.T) {
	cs := newTestCryptoService()
	sig := cs.MakeXSignatureBasic("POST", "api/v8/info", 1700000000)
	if sig == "" {
		t.Fatal("Expected non-empty basic signature")
	}
}
