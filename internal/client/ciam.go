package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kertasbaru/me-cli-sunset/internal/crypto"
)

// RequestOTP sends an OTP to the specified phone number.
func (c *Client) RequestOTP(contact string) (string, error) {
	reqURL := c.baseCIAMURL + "/realms/xl-ciam/auth/otp"

	params := url.Values{}
	params.Set("contact", contact)
	params.Set("contactType", "SMS")
	params.Set("alternateContact", "false")

	gmt7 := time.FixedZone("GMT+7", 7*3600)
	now := time.Now().In(gmt7)

	req, err := http.NewRequest(http.MethodGet, reqURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setCIAMHeaders(req, now)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	subscriberID, ok := result["subscriber_id"].(string)
	if !ok {
		errMsg, _ := result["error"].(string)
		return "", fmt.Errorf("subscriber_id not found in response: %s", errMsg)
	}

	return subscriberID, nil
}

// SubmitOTP verifies the OTP code and returns authentication tokens.
func (c *Client) SubmitOTP(contactType, contact, code string) (map[string]interface{}, error) {
	reqURL := c.baseCIAMURL + "/realms/xl-ciam/protocol/openid-connect/token"

	finalContact := contact
	finalCode := code

	if contactType == "DEVICEID" {
		finalContact = base64.StdEncoding.EncodeToString([]byte(contact))
	}

	gmt7 := time.FixedZone("GMT+7", 7*3600)
	nowGMT7 := time.Now().In(gmt7)

	tsForSign := crypto.TSWithoutColon(nowGMT7)
	tsHeader := crypto.TSWithoutColon(nowGMT7.Add(-5 * time.Minute))
	signature := c.crypto.MakeAXAPISignature(tsForSign, finalContact, code, contactType)

	payload := fmt.Sprintf("contactType=%s&code=%s&grant_type=password&contact=%s&scope=openid",
		contactType, finalCode, finalContact)

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Authorization", "Basic "+c.basicAuth)
	req.Header.Set("Ax-Api-Signature", signature)
	req.Header.Set("Ax-Device-Id", c.deviceID)
	req.Header.Set("Ax-Fingerprint", c.fingerprint)
	req.Header.Set("Ax-Request-At", tsHeader)
	req.Header.Set("Ax-Request-Device", "samsung")
	req.Header.Set("Ax-Request-Device-Model", "SM-N935F")
	req.Header.Set("Ax-Request-Id", uuid.New().String())
	req.Header.Set("Ax-Substype", "PREPAID")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if errVal, ok := result["error"]; ok {
		return nil, fmt.Errorf("OTP submission failed: %v", errVal)
	}

	return result, nil
}

// RefreshToken gets new tokens using a refresh token.
func (c *Client) RefreshToken(refreshToken, subscriberID string) (map[string]interface{}, error) {
	reqURL := c.baseCIAMURL + "/realms/xl-ciam/protocol/openid-connect/token"

	gmt7 := time.FixedZone("GMT+7", 7*3600)
	now := time.Now().In(gmt7)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	host := strings.TrimPrefix(c.baseCIAMURL, "https://")
	req.Header.Set("Host", host)
	req.Header.Set("Ax-Request-At", now.Format("2006-01-02T15:04:05.000")+"+0700")
	req.Header.Set("Ax-Device-Id", c.deviceID)
	req.Header.Set("Ax-Request-Id", uuid.New().String())
	req.Header.Set("Ax-Request-Device", "samsung")
	req.Header.Set("Ax-Request-Device-Model", "SM-N935F")
	req.Header.Set("Ax-Fingerprint", c.fingerprint)
	req.Header.Set("Authorization", "Basic "+c.basicAuth)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Ax-Substype", "PREPAID")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 400 {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err == nil {
			if errDesc, ok := errResp["error_description"].(string); ok && errDesc == "Session not active" {
				return c.handleSessionExtension(subscriberID)
			}
		}
		return nil, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if _, ok := result["id_token"]; !ok {
		return nil, fmt.Errorf("id_token not found in response")
	}

	return result, nil
}

// ExtendSession extends an expired session using the subscriber ID.
func (c *Client) ExtendSession(subscriberID string) (string, error) {
	b64SubID := base64.StdEncoding.EncodeToString([]byte(subscriberID))
	reqURL := c.baseCIAMURL + "/realms/xl-ciam/auth/extend-session"

	params := url.Values{}
	params.Set("contact", b64SubID)
	params.Set("contactType", "DEVICEID")

	gmt7 := time.FixedZone("GMT+7", 7*3600)
	now := time.Now().In(gmt7)

	req, err := http.NewRequest(http.MethodGet, reqURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setCIAMHeaders(req, now)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("extend session failed with status %d (failed to read body: %w)", resp.StatusCode, err)
		}
		return "", fmt.Errorf("extend session failed: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("data field not found in response")
	}

	exchangeCode, ok := data["exchange_code"].(string)
	if !ok {
		return "", fmt.Errorf("exchange_code not found in response")
	}

	return exchangeCode, nil
}

// GetAuthCode generates an authorization code for balance transfer.
func (c *Client) GetAuthCode(accessToken, pin, msisdn string) (string, error) {
	reqURL := c.baseCIAMURL + "/ciam/auth/authorization-token/generate"

	host := strings.TrimPrefix(c.baseCIAMURL, "https://")
	gmt7 := time.FixedZone("GMT+7", 7*3600)
	now := time.Now().In(gmt7)

	pinB64 := base64.StdEncoding.EncodeToString([]byte(pin))

	payload := map[string]interface{}{
		"pin":              pinB64,
		"transaction_type": "SHARE_BALANCE",
		"receiver_msisdn":  msisdn,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(string(payloadJSON)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Host", host)
	req.Header.Set("Ax-Request-At", now.Format("2006-01-02T15:04:05.000")+"+0700")
	req.Header.Set("Ax-Device-Id", c.deviceID)
	req.Header.Set("Ax-Request-Id", uuid.New().String())
	req.Header.Set("Ax-Request-Device", "samsung")
	req.Header.Set("Ax-Request-Device-Model", "SM-N935F")
	req.Header.Set("Ax-Fingerprint", c.fingerprint)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Ax-Substype", "PREPAID")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("get auth code failed: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	status, _ := result["status"].(string)
	if status != "Success" {
		return "", fmt.Errorf("auth code request failed with status: %s", status)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("data field not found")
	}

	authCode, ok := data["authorization_code"].(string)
	if !ok {
		return "", fmt.Errorf("authorization_code not found")
	}

	return authCode, nil
}

func (c *Client) handleSessionExtension(subscriberID string) (map[string]interface{}, error) {
	if subscriberID == "" {
		return nil, fmt.Errorf("subscriber ID is required for session extension")
	}

	exchangeCode, err := c.ExtendSession(subscriberID)
	if err != nil {
		return nil, fmt.Errorf("failed to extend session: %w", err)
	}

	return c.SubmitOTP("DEVICEID", subscriberID, exchangeCode)
}

func (c *Client) setCIAMHeaders(req *http.Request, now time.Time) {
	host := strings.TrimPrefix(c.baseCIAMURL, "https://")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Authorization", "Basic "+c.basicAuth)
	req.Header.Set("Ax-Device-Id", c.deviceID)
	req.Header.Set("Ax-Fingerprint", c.fingerprint)
	req.Header.Set("Ax-Request-At", crypto.JavaLikeTimestamp(now))
	req.Header.Set("Ax-Request-Device", "samsung")
	req.Header.Set("Ax-Request-Device-Model", "SM-N935F")
	req.Header.Set("Ax-Request-Id", uuid.New().String())
	req.Header.Set("Ax-Substype", "PREPAID")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", host)
	req.Header.Set("User-Agent", c.userAgent)
}
