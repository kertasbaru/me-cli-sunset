// Package crypto provides cryptographic operations for API communication.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CryptoService provides encryption, decryption, and signing operations.
type CryptoService struct {
	XDataKey          string
	AXAPISigKey       string
	XAPIBaseSecret    string
	EncryptedFieldKey string
	AXFPKey           string
}

// NewCryptoService creates a new CryptoService with the given keys.
func NewCryptoService(xdataKey, axAPISigKey, xAPIBaseSecret, encryptedFieldKey, axFPKey string) *CryptoService {
	return &CryptoService{
		XDataKey:          xdataKey,
		AXAPISigKey:       axAPISigKey,
		XAPIBaseSecret:    xAPIBaseSecret,
		EncryptedFieldKey: encryptedFieldKey,
		AXFPKey:           axFPKey,
	}
}

// --- PKCS7 Padding ---

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padBytes := make([]byte, padding)
	for i := range padBytes {
		padBytes[i] = byte(padding)
	}
	return append(data, padBytes...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding byte")
		}
	}
	return data[:len(data)-padding], nil
}

// --- XData Encryption/Decryption ---

func deriveIV(xtimeMs int64) []byte {
	hash := sha256.Sum256([]byte(strconv.FormatInt(xtimeMs, 10)))
	hexStr := hex.EncodeToString(hash[:])
	return []byte(hexStr[:16])
}

// EncryptXData encrypts plaintext using AES-CBC with a time-derived IV.
func (cs *CryptoService) EncryptXData(plaintext string, xtimeMs int64) (string, error) {
	iv := deriveIV(xtimeMs)
	key := []byte(cs.XDataKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	padded := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(ciphertext), nil
}

// DecryptXData decrypts xdata using AES-CBC with a time-derived IV.
func (cs *CryptoService) DecryptXData(xdata string, xtimeMs int64) (string, error) {
	iv := deriveIV(xtimeMs)
	key := []byte(cs.XDataKey)

	// Add padding if needed
	if m := len(xdata) % 4; m != 0 {
		xdata += strings.Repeat("=", 4-m)
	}

	ciphertext, err := base64.URLEncoding.DecodeString(xdata)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of block size")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to unpad: %w", err)
	}

	return string(unpadded), nil
}

// --- HMAC Signatures ---

// MakeXSignature creates an x-signature for standard API requests.
func (cs *CryptoService) MakeXSignature(idToken, method, path string, sigTimeSec int64) string {
	keyStr := fmt.Sprintf("%s;%s;%s;%s;%d", cs.XAPIBaseSecret, idToken, method, path, sigTimeSec)
	msg := fmt.Sprintf("%s;%d;", idToken, sigTimeSec)
	return cs.hmacSHA512Hex(keyStr, msg)
}

// MakeXSignaturePayment creates an x-signature for payment requests.
func (cs *CryptoService) MakeXSignaturePayment(accessToken string, sigTimeSec int64, packageCode, tokenPayment, paymentMethod, paymentFor, path string) string {
	keyStr := fmt.Sprintf("%s;%d#ae-hei_9Tee6he+Ik3Gais5=;POST;%s;%d", cs.XAPIBaseSecret, sigTimeSec, path, sigTimeSec)
	msg := fmt.Sprintf("%s;%s;%d;%s;%s;%s;", accessToken, tokenPayment, sigTimeSec, paymentFor, paymentMethod, packageCode)
	return cs.hmacSHA512Hex(keyStr, msg)
}

// MakeAXAPISignature creates an ax-api-signature for CIAM requests.
func (cs *CryptoService) MakeAXAPISignature(tsForSign, contact, code, contactType string) string {
	key := []byte(cs.AXAPISigKey)
	preimage := fmt.Sprintf("%spassword%s%s%sopenid", tsForSign, contactType, contact, code)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(preimage))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// MakeXSignatureBounty creates an x-signature for bounty exchange requests.
func (cs *CryptoService) MakeXSignatureBounty(accessToken string, sigTimeSec int64, packageCode, tokenPayment string) string {
	path := "api/v8/personalization/bounties-exchange"
	keyStr := fmt.Sprintf("%s;%s;%d#ae-hei_9Tee6he+Ik3Gais5=;POST;%s;%d", cs.XAPIBaseSecret, accessToken, sigTimeSec, path, sigTimeSec)
	msg := fmt.Sprintf("%s;%s;%d;%s;", accessToken, tokenPayment, sigTimeSec, packageCode)
	return cs.hmacSHA512Hex(keyStr, msg)
}

// MakeXSignatureLoyalty creates an x-signature for loyalty redemption requests.
func (cs *CryptoService) MakeXSignatureLoyalty(sigTimeSec int64, packageCode, tokenConfirmation, path string) string {
	keyStr := fmt.Sprintf("%s;%d#ae-hei_9Tee6he+Ik3Gais5=;POST;%s;%d", cs.XAPIBaseSecret, sigTimeSec, path, sigTimeSec)
	msg := fmt.Sprintf("%s;%d;%s;", tokenConfirmation, sigTimeSec, packageCode)
	return cs.hmacSHA512Hex(keyStr, msg)
}

// MakeXSignatureBountyAllotment creates an x-signature for bounty allotment requests.
func (cs *CryptoService) MakeXSignatureBountyAllotment(sigTimeSec int64, packageCode, tokenConfirmation, path, destinationMSISDN string) string {
	keyStr := fmt.Sprintf("%s;%d#ae-hei_9Tee6he+Ik3Gais5=;%s;POST;%s;%d", cs.XAPIBaseSecret, sigTimeSec, destinationMSISDN, path, sigTimeSec)
	msg := fmt.Sprintf("%s;%d;%s;%s;", tokenConfirmation, sigTimeSec, destinationMSISDN, packageCode)
	return cs.hmacSHA512Hex(keyStr, msg)
}

// MakeXSignatureBasic creates a basic x-signature for unauthenticated requests.
func (cs *CryptoService) MakeXSignatureBasic(method, path string, sigTimeSec int64) string {
	keyStr := fmt.Sprintf("%s;%s;%s;%d", cs.XAPIBaseSecret, method, path, sigTimeSec)
	msg := fmt.Sprintf("%d;en;", sigTimeSec)
	return cs.hmacSHA512Hex(keyStr, msg)
}

func (cs *CryptoService) hmacSHA512Hex(key, msg string) string {
	mac := hmac.New(sha512.New, []byte(key))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

// --- Encrypted Field ---

// BuildEncryptedField generates an encrypted field value with random IV.
func (cs *CryptoService) BuildEncryptedField(urlsafe bool) (string, error) {
	key := []byte(cs.EncryptedFieldKey)

	ivBytes := make([]byte, 8)
	if _, err := rand.Read(ivBytes); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}
	ivHex := hex.EncodeToString(ivBytes)
	iv := []byte(ivHex)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	padded := pkcs7Pad([]byte{}, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	var encoded string
	if urlsafe {
		encoded = base64.URLEncoding.EncodeToString(ciphertext)
	} else {
		encoded = base64.StdEncoding.EncodeToString(ciphertext)
	}

	return encoded + ivHex, nil
}

// --- Circle MSISDN Encryption ---

// EncryptCircleMSISDN encrypts a phone number for circle member operations.
func (cs *CryptoService) EncryptCircleMSISDN(msisdn string) (string, error) {
	key := []byte(cs.EncryptedFieldKey)

	ivBytes := make([]byte, 8)
	if _, err := rand.Read(ivBytes); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}
	ivHex := hex.EncodeToString(ivBytes)
	iv := []byte(ivHex)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	padded := pkcs7Pad([]byte(msisdn), aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(ciphertext)
	return encoded + ivHex, nil
}

// DecryptCircleMSISDN decrypts an encrypted phone number from circle operations.
func (cs *CryptoService) DecryptCircleMSISDN(encrypted string) (string, error) {
	if len(encrypted) < 16 {
		return "", fmt.Errorf("encrypted string too short")
	}

	ivASCII := encrypted[len(encrypted)-16:]
	b64Part := encrypted[:len(encrypted)-16]

	key := []byte(cs.EncryptedFieldKey)
	iv := []byte(ivASCII)

	if m := len(b64Part) % 4; m != 0 {
		b64Part += strings.Repeat("=", 4-m)
	}

	ciphertext, err := base64.URLEncoding.DecodeString(b64Part)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to unpad: %w", err)
	}

	return string(unpadded), nil
}

// --- Device Fingerprint ---

// DeviceInfo holds device information for fingerprint generation.
type DeviceInfo struct {
	Manufacturer   string
	Model          string
	Lang           string
	Resolution     string
	TZShort        string
	IP             string
	FontScale      float64
	AndroidRelease string
	MSISDN         string
}

// GenerateFingerprint creates an AES-encrypted device fingerprint.
func (cs *CryptoService) GenerateFingerprint(dev DeviceInfo) (string, error) {
	plain := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%g|Android %s|%s",
		dev.Manufacturer, dev.Model, dev.Lang, dev.Resolution,
		dev.TZShort, dev.IP, dev.FontScale, dev.AndroidRelease, dev.MSISDN)

	key := []byte(cs.AXFPKey)
	iv := make([]byte, 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	padded := pkcs7Pad([]byte(plain), aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DeviceIDFromFingerprint generates an MD5-based device ID from a fingerprint.
func DeviceIDFromFingerprint(fingerprint string) string {
	hash := md5.Sum([]byte(fingerprint))
	return hex.EncodeToString(hash[:])
}

// --- Timestamp Utilities ---

// JavaLikeTimestamp formats a time in Java-compatible ISO format with timezone.
func JavaLikeTimestamp(t time.Time) string {
	ms := t.Nanosecond() / 10000000
	tz := t.Format("-0700")
	tzColon := tz[:3] + ":" + tz[3:]
	return fmt.Sprintf("%s.%02d%s", t.Format("2006-01-02T15:04:05"), ms, tzColon)
}

// TSWithoutColon formats a time in GMT+7 with milliseconds but without colon in timezone.
func TSWithoutColon(t time.Time) string {
	gmt7 := time.FixedZone("GMT+7", 7*3600)
	t = t.In(gmt7)
	ms := t.Nanosecond() / 1000000
	tz := t.Format("-0700")
	return fmt.Sprintf("%s.%03d%s", t.Format("2006-01-02T15:04:05"), ms, tz)
}
