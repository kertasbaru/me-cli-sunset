// Package model defines data structures used across the application.
package model

import "time"

// Account represents a user account stored in the database.
type Account struct {
	ID               int64     `json:"id"`
	Number           int64     `json:"number"`
	SubscriberID     string    `json:"subscriber_id"`
	SubscriptionType string    `json:"subscription_type"`
	RefreshToken     string    `json:"refresh_token"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Bookmark represents a saved package bookmark.
type Bookmark struct {
	ID           int64     `json:"id"`
	AccountID    int64     `json:"account_id"`
	FamilyCode   string    `json:"family_code"`
	FamilyName   string    `json:"family_name"`
	IsEnterprise bool      `json:"is_enterprise"`
	VariantName  string    `json:"variant_name"`
	OptionName   string    `json:"option_name"`
	Order        int       `json:"order"`
	CreatedAt    time.Time `json:"created_at"`
}

// DeviceFingerprint stores generated device fingerprints.
type DeviceFingerprint struct {
	ID          int64     `json:"id"`
	Fingerprint string    `json:"fingerprint"`
	DeviceID    string    `json:"device_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// Tokens holds the authentication tokens for a session.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}

// ActiveUser represents the currently active user with session tokens.
type ActiveUser struct {
	Number           int64  `json:"number"`
	SubscriberID     string `json:"subscriber_id"`
	SubscriptionType string `json:"subscription_type"`
	Tokens           Tokens `json:"tokens"`
}

// PaymentItem represents an item in a purchase settlement.
type PaymentItem struct {
	PackageOptionCode  string `json:"package_option_code"`
	PackageFamilyCode  string `json:"package_family_code"`
	PackageVariantCode string `json:"package_variant_code"`
	TokenConfirmation  string `json:"token_confirmation"`
	Amount             int    `json:"amount"`
}

// --- API Request/Response Types ---

// OTPRequest represents a request to send an OTP.
type OTPRequest struct {
	Contact string `json:"contact"`
}

// OTPSubmitRequest represents a request to verify an OTP.
type OTPSubmitRequest struct {
	ContactType string `json:"contact_type"`
	Contact     string `json:"contact"`
	Code        string `json:"code"`
}

// PackageRequest represents a request to get package details.
type PackageRequest struct {
	PackageOptionCode  string `json:"package_option_code"`
	PackageFamilyCode  string `json:"package_family_code,omitempty"`
	PackageVariantCode string `json:"package_variant_code,omitempty"`
}

// FamilyRequest represents a request to get family package data.
type FamilyRequest struct {
	FamilyCode    string `json:"family_code"`
	IsEnterprise  *bool  `json:"is_enterprise,omitempty"`
	MigrationType string `json:"migration_type,omitempty"`
}

// PurchaseBalanceRequest represents a purchase using balance payment.
type PurchaseBalanceRequest struct {
	Items         []PaymentItem `json:"items"`
	PaymentFor    string        `json:"payment_for"`
	TopupNumber   string        `json:"topup_number,omitempty"`
	OverridePrice int           `json:"override_price,omitempty"`
}

// PurchaseQRISRequest represents a purchase using QRIS payment.
type PurchaseQRISRequest struct {
	Items      []PaymentItem `json:"items"`
	PaymentFor string        `json:"payment_for"`
}

// PurchaseEWalletRequest represents a purchase using e-wallet payment.
type PurchaseEWalletRequest struct {
	Items         []PaymentItem `json:"items"`
	PaymentFor    string        `json:"payment_for"`
	PaymentMethod string        `json:"payment_method"`
}

// CircleCreateRequest represents a request to create a family circle.
type CircleCreateRequest struct {
	ParentName   string `json:"parent_name"`
	GroupName    string `json:"group_name"`
	MemberMSISDN string `json:"member_msisdn"`
	MemberName   string `json:"member_name"`
}

// CircleMemberRequest represents a request to manage circle members.
type CircleMemberRequest struct {
	MSISDN         string `json:"msisdn"`
	Name           string `json:"name,omitempty"`
	GroupID        string `json:"group_id"`
	MemberIDParent string `json:"member_id_parent,omitempty"`
}

// FamplanQuotaRequest represents a request to set family plan quota.
type FamplanQuotaRequest struct {
	FamilyMemberID     string `json:"family_member_id"`
	OriginalAllocation int    `json:"original_allocation"`
	NewAllocation      int    `json:"new_allocation"`
}

// BookmarkRequest represents a request to add a bookmark.
type BookmarkRequest struct {
	FamilyCode   string `json:"family_code"`
	FamilyName   string `json:"family_name"`
	IsEnterprise bool   `json:"is_enterprise"`
	VariantName  string `json:"variant_name"`
	OptionName   string `json:"option_name"`
	Order        int    `json:"order"`
}

// APIResponse is a generic API response wrapper.
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
