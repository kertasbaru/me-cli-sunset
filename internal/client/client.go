// Package client provides HTTP client operations for communicating with the upstream API.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kertasbaru/me-cli-sunset/internal/crypto"
)

// Client handles encrypted API communication with the upstream server.
type Client struct {
	baseAPIURL  string
	baseCIAMURL string
	apiKey      string
	basicAuth   string
	userAgent   string
	crypto      *crypto.CryptoService
	httpClient  *http.Client
	fingerprint string
	deviceID    string
}

// NewClient creates a new API client.
func NewClient(baseAPIURL, baseCIAMURL, apiKey, basicAuth, userAgent string, cs *crypto.CryptoService, fingerprint, deviceID string) *Client {
	return &Client{
		baseAPIURL:  baseAPIURL,
		baseCIAMURL: baseCIAMURL,
		apiKey:      apiKey,
		basicAuth:   basicAuth,
		userAgent:   userAgent,
		crypto:      cs,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		fingerprint: fingerprint,
		deviceID:    deviceID,
	}
}

// SendAPIRequest sends an encrypted and signed request to the upstream API.
func (c *Client) SendAPIRequest(path string, payload map[string]interface{}, idToken, method string) (map[string]interface{}, error) {
	plainBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	xtimeMs := time.Now().UnixMilli()
	xdata, err := c.crypto.EncryptXData(string(plainBody), xtimeMs)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt xdata: %w", err)
	}

	sigTimeSec := xtimeMs / 1000
	xSig := c.crypto.MakeXSignature(idToken, method, path, sigTimeSec)

	body := map[string]interface{}{
		"xdata": xdata,
		"xtime": xtimeMs,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	now := time.Now()
	host := strings.TrimPrefix(c.baseAPIURL, "https://")

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", c.baseAPIURL, path), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Host", host)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+idToken)
	req.Header.Set("X-Hv", "v3")
	req.Header.Set("X-Signature-Time", fmt.Sprintf("%d", sigTimeSec))
	req.Header.Set("X-Signature", xSig)
	req.Header.Set("X-Request-Id", uuid.New().String())
	req.Header.Set("X-Request-At", crypto.JavaLikeTimestamp(now))
	req.Header.Set("X-Version-App", "8.9.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var encryptedResp map[string]interface{}
	if err := json.Unmarshal(respBody, &encryptedResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	xdataResp, ok := encryptedResp["xdata"].(string)
	if !ok {
		return encryptedResp, nil
	}

	xtimeResp, ok := encryptedResp["xtime"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing xtime in response")
	}

	decrypted, err := c.crypto.DecryptXData(xdataResp, int64(xtimeResp))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(decrypted), &result); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted response: %w", err)
	}

	return result, nil
}
