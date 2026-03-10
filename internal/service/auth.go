// Package service provides business logic services.
package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kertasbaru/me-cli-sunset/internal/client"
	"github.com/kertasbaru/me-cli-sunset/internal/database"
	"github.com/kertasbaru/me-cli-sunset/internal/model"
)

// tokenRefreshIntervalSec is the interval in seconds before tokens are automatically refreshed.
const tokenRefreshIntervalSec = 300

// AuthService manages authentication state and token lifecycle.
type AuthService struct {
	db              *database.DB
	client          *client.Client
	apiKey          string
	activeUser      *model.ActiveUser
	lastRefreshTime int64
	mu              sync.RWMutex
}

// NewAuthService creates a new AuthService.
func NewAuthService(db *database.DB, cl *client.Client, apiKey string) *AuthService {
	svc := &AuthService{
		db:              db,
		client:          cl,
		apiKey:          apiKey,
		lastRefreshTime: time.Now().Unix(),
	}

	// Try to load active account from database
	account, err := db.GetActiveAccount()
	if err != nil {
		log.Printf("Failed to load active account: %v", err)
	} else if account != nil {
		if err := svc.loadActiveUser(account); err != nil {
			log.Printf("Failed to initialize active user: %v", err)
		}
	}

	return svc
}

// GetActiveUser returns the current active user, refreshing tokens if needed.
func (s *AuthService) GetActiveUser() *model.ActiveUser {
	s.mu.RLock()
	user := s.activeUser
	lastRefresh := s.lastRefreshTime
	s.mu.RUnlock()

	if user == nil {
		accounts, err := s.db.GetAllAccounts()
		if err != nil || len(accounts) == 0 {
			return nil
		}
		if err := s.loadActiveUser(&accounts[0]); err != nil {
			return nil
		}
		s.mu.RLock()
		user = s.activeUser
		s.mu.RUnlock()
	}

	// Refresh tokens if older than the configured interval
	if time.Now().Unix()-lastRefresh > tokenRefreshIntervalSec {
		if err := s.renewTokens(); err != nil {
			log.Printf("Failed to renew tokens: %v", err)
		}
	}

	return user
}

// GetActiveTokens returns the current active user's tokens.
func (s *AuthService) GetActiveTokens() *model.Tokens {
	user := s.GetActiveUser()
	if user == nil {
		return nil
	}
	return &user.Tokens
}

// AddAccount adds or updates an account with the given refresh token.
func (s *AuthService) AddAccount(number int64, refreshToken string) error {
	tokens, err := s.client.RefreshToken(refreshToken, "")
	if err != nil {
		return fmt.Errorf("failed to get tokens: %w", err)
	}

	accessToken, _ := tokens["access_token"].(string)
	idToken, _ := tokens["id_token"].(string)
	newRefreshToken, _ := tokens["refresh_token"].(string)

	profileData, err := s.client.GetProfile(accessToken, idToken)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile, _ := profileData["profile"].(map[string]interface{})
	subscriberID, _ := profile["subscriber_id"].(string)
	subscriptionType, _ := profile["subscription_type"].(string)

	if err := s.db.UpsertAccount(number, subscriberID, subscriptionType, newRefreshToken); err != nil {
		return fmt.Errorf("failed to save account: %w", err)
	}

	return s.SetActiveAccount(number)
}

// RemoveAccount removes an account by phone number.
func (s *AuthService) RemoveAccount(number int64) error {
	if err := s.db.DeleteAccount(number); err != nil {
		return err
	}

	s.mu.Lock()
	if s.activeUser != nil && s.activeUser.Number == number {
		s.activeUser = nil
	}
	s.mu.Unlock()

	// Switch to first available account
	accounts, err := s.db.GetAllAccounts()
	if err != nil {
		return err
	}
	if len(accounts) > 0 {
		return s.SetActiveAccount(accounts[0].Number)
	}

	return nil
}

// SetActiveAccount sets the specified account as the active one.
func (s *AuthService) SetActiveAccount(number int64) error {
	if err := s.db.SetActiveAccount(number); err != nil {
		return err
	}

	account, err := s.db.GetAccountByNumber(number)
	if err != nil || account == nil {
		return fmt.Errorf("account not found: %d", number)
	}

	return s.loadActiveUser(account)
}

// GetAllAccounts returns all stored accounts.
func (s *AuthService) GetAllAccounts() ([]model.Account, error) {
	return s.db.GetAllAccounts()
}

func (s *AuthService) loadActiveUser(account *model.Account) error {
	tokens, err := s.client.RefreshToken(account.RefreshToken, account.SubscriberID)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	accessToken, _ := tokens["access_token"].(string)
	idToken, _ := tokens["id_token"].(string)
	newRefreshToken, _ := tokens["refresh_token"].(string)

	profileData, err := s.client.GetProfile(accessToken, idToken)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile, _ := profileData["profile"].(map[string]interface{})
	subscriberID, _ := profile["subscriber_id"].(string)
	subscriptionType, _ := profile["subscription_type"].(string)

	// Update stored account data
	if err := s.db.UpsertAccount(account.Number, subscriberID, subscriptionType, newRefreshToken); err != nil {
		log.Printf("Failed to update account: %v", err)
	}
	if err := s.db.SetActiveAccount(account.Number); err != nil {
		log.Printf("Failed to set active account: %v", err)
	}

	s.mu.Lock()
	s.activeUser = &model.ActiveUser{
		Number:           account.Number,
		SubscriberID:     subscriberID,
		SubscriptionType: subscriptionType,
		Tokens: model.Tokens{
			AccessToken:  accessToken,
			IDToken:      idToken,
			RefreshToken: newRefreshToken,
		},
	}
	s.lastRefreshTime = time.Now().Unix()
	s.mu.Unlock()

	return nil
}

func (s *AuthService) renewTokens() error {
	s.mu.RLock()
	user := s.activeUser
	s.mu.RUnlock()

	if user == nil {
		return fmt.Errorf("no active user")
	}

	tokens, err := s.client.RefreshToken(user.Tokens.RefreshToken, user.SubscriberID)
	if err != nil {
		return err
	}

	accessToken, _ := tokens["access_token"].(string)
	idToken, _ := tokens["id_token"].(string)
	newRefreshToken, _ := tokens["refresh_token"].(string)

	if err := s.db.UpsertAccount(user.Number, user.SubscriberID, user.SubscriptionType, newRefreshToken); err != nil {
		log.Printf("Failed to update refresh token: %v", err)
	}

	s.mu.Lock()
	s.activeUser.Tokens = model.Tokens{
		AccessToken:  accessToken,
		IDToken:      idToken,
		RefreshToken: newRefreshToken,
	}
	s.lastRefreshTime = time.Now().Unix()
	s.mu.Unlock()

	return nil
}
