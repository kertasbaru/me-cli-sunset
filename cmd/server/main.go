// Package main is the entry point for the ME CLI Sunset API server.
//
// This server provides a RESTful API for managing XL Axiata mobile accounts,
// packages, family plans, and circle groups. It replaces the original Python CLI
// with a professional Go API service using SQLite for credential storage.
//
// Repository: https://github.com/kertasbaru/me-cli-sunset
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/kertasbaru/me-cli-sunset/internal/client"
	"github.com/kertasbaru/me-cli-sunset/internal/config"
	"github.com/kertasbaru/me-cli-sunset/internal/crypto"
	"github.com/kertasbaru/me-cli-sunset/internal/database"
	"github.com/kertasbaru/me-cli-sunset/internal/handler"
	"github.com/kertasbaru/me-cli-sunset/internal/middleware"
	"github.com/kertasbaru/me-cli-sunset/internal/service"
)

func main() {
	// Load .env file if present
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized successfully")

	// Initialize crypto service
	cryptoSvc := crypto.NewCryptoService(
		cfg.XDataKey,
		cfg.AXAPISigKey,
		cfg.XAPIBaseSecret,
		cfg.EncryptedFieldKey,
		cfg.AXFPKey,
	)

	// Initialize or load device fingerprint
	fingerprint, deviceID, err := initFingerprint(db, cryptoSvc)
	if err != nil {
		log.Fatalf("Failed to initialize fingerprint: %v", err)
	}

	// Initialize API client
	apiClient := client.NewClient(
		cfg.BaseAPIURL,
		cfg.BaseCIAMURL,
		cfg.APIKey,
		cfg.BasicAuth,
		cfg.UserAgent,
		cryptoSvc,
		fingerprint,
		deviceID,
	)

	// Initialize auth service
	authSvc := service.NewAuthService(db, apiClient, cfg.APIKey)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authSvc, apiClient)
	pkgHandler := handler.NewPackageHandler(authSvc, apiClient)
	circleHandler := handler.NewCircleHandler(authSvc, apiClient)
	familyHandler := handler.NewFamilyHandler(authSvc, apiClient)
	notifHandler := handler.NewNotificationHandler(authSvc, apiClient)
	bookmarkHandler := handler.NewBookmarkHandler(authSvc, db)

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"me-cli-sunset-api","version":"1.0.0"}`))
	})

	// Auth routes
	mux.HandleFunc("POST /api/v1/auth/otp/request", authHandler.RequestOTP)
	mux.HandleFunc("POST /api/v1/auth/otp/submit", authHandler.SubmitOTP)
	mux.HandleFunc("GET /api/v1/auth/accounts", authHandler.ListAccounts)
	mux.HandleFunc("POST /api/v1/auth/accounts/switch", authHandler.SwitchAccount)
	mux.HandleFunc("DELETE /api/v1/auth/accounts/{number}", authHandler.DeleteAccount)
	mux.HandleFunc("GET /api/v1/auth/active", authHandler.GetActiveUser)
	mux.HandleFunc("POST /api/v1/auth/renew", authHandler.RenewToken)

	// Package routes
	mux.HandleFunc("GET /api/v1/packages/balance", pkgHandler.GetBalance)
	mux.HandleFunc("POST /api/v1/packages/family", pkgHandler.GetFamily)
	mux.HandleFunc("GET /api/v1/packages/families", pkgHandler.GetFamilies)
	mux.HandleFunc("POST /api/v1/packages/detail", pkgHandler.GetPackageDetail)
	mux.HandleFunc("GET /api/v1/packages/addons", pkgHandler.GetAddons)
	mux.HandleFunc("GET /api/v1/packages/tiering", pkgHandler.GetTieringInfo)
	mux.HandleFunc("GET /api/v1/packages/quotas", pkgHandler.GetQuotaDetails)
	mux.HandleFunc("POST /api/v1/packages/payment-methods", pkgHandler.GetPaymentMethods)
	mux.HandleFunc("POST /api/v1/packages/unsubscribe", pkgHandler.Unsubscribe)

	// Circle routes
	mux.HandleFunc("GET /api/v1/circle/groups", circleHandler.GetGroupData)
	mux.HandleFunc("POST /api/v1/circle/groups", circleHandler.CreateCircle)
	mux.HandleFunc("GET /api/v1/circle/groups/{group_id}/members", circleHandler.GetGroupMembers)
	mux.HandleFunc("POST /api/v1/circle/members/validate", circleHandler.ValidateMember)
	mux.HandleFunc("POST /api/v1/circle/members/invite", circleHandler.InviteMember)
	mux.HandleFunc("POST /api/v1/circle/members/remove", circleHandler.RemoveMember)
	mux.HandleFunc("POST /api/v1/circle/invitations/accept", circleHandler.AcceptInvitation)
	mux.HandleFunc("POST /api/v1/circle/spending-tracker", circleHandler.GetSpendingTracker)
	mux.HandleFunc("POST /api/v1/circle/bonus", circleHandler.GetBonusData)

	// Family Plan routes
	mux.HandleFunc("GET /api/v1/family/plan", familyHandler.GetFamilyPlanData)
	mux.HandleFunc("POST /api/v1/family/validate", familyHandler.ValidateMSISDN)
	mux.HandleFunc("POST /api/v1/family/members/change", familyHandler.ChangeMember)
	mux.HandleFunc("POST /api/v1/family/members/remove", familyHandler.RemoveMember)
	mux.HandleFunc("POST /api/v1/family/quota", familyHandler.SetQuotaLimit)

	// Notification & Transaction routes
	mux.HandleFunc("GET /api/v1/notifications", notifHandler.GetNotifications)
	mux.HandleFunc("GET /api/v1/notifications/{notification_id}", notifHandler.GetNotificationDetail)
	mux.HandleFunc("GET /api/v1/transactions", notifHandler.GetTransactionHistory)
	mux.HandleFunc("GET /api/v1/dashboard", notifHandler.GetDashboard)

	// Bookmark routes
	mux.HandleFunc("GET /api/v1/bookmarks", bookmarkHandler.ListBookmarks)
	mux.HandleFunc("POST /api/v1/bookmarks", bookmarkHandler.AddBookmark)
	mux.HandleFunc("DELETE /api/v1/bookmarks/{id}", bookmarkHandler.DeleteBookmark)

	// API documentation routes
	openapiSpec, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		log.Printf("Warning: could not load api/openapi.yaml: %v", err)
		openapiSpec = []byte("openapi: 3.0.3\ninfo:\n  title: ME CLI Sunset API\n  version: 1.0.0\npaths: {}")
	}
	mux.HandleFunc("GET /docs", handler.DocsPage("/api/v1/openapi.yaml"))
	mux.HandleFunc("GET /api/v1/openapi.yaml", handler.ServeOpenAPISpec(openapiSpec))

	// Apply middleware
	var h http.Handler = mux
	h = middleware.Recovery(h)
	h = middleware.CORS(h)
	h = middleware.Logger(h)

	addr := cfg.ServerHost + ":" + cfg.ServerPort
	log.Printf("Starting ME CLI Sunset API server on %s", addr)
	log.Printf("API documentation: http://%s/docs", addr)

	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

const (
	// randModelMin and randModelMax define the range for random device model generation.
	randModelMin = 1000
	randModelMax = 9000
)

func initFingerprint(db *database.DB, cs *crypto.CryptoService) (string, string, error) {
	existing, err := db.GetFingerprint()
	if err != nil {
		return "", "", fmt.Errorf("failed to query fingerprint: %w", err)
	}

	if existing != nil {
		return existing.Fingerprint, existing.DeviceID, nil
	}

	// Generate new fingerprint
	dev := crypto.DeviceInfo{
		Manufacturer:   fmt.Sprintf("samsung%d", rand.Intn(randModelMax)+randModelMin),
		Model:          fmt.Sprintf("SM-N93%d", rand.Intn(randModelMax)+randModelMin),
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
		return "", "", fmt.Errorf("failed to generate fingerprint: %w", err)
	}

	deviceID := crypto.DeviceIDFromFingerprint(fp)

	if err := db.SaveFingerprint(fp, deviceID); err != nil {
		// Log but don't fail - fingerprint can be re-generated
		log.Printf("Warning: failed to save fingerprint: %v", err)
	}

	return fp, deviceID, nil
}

func init() {
	// Ensure usage info if run without args and env
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("ME CLI Sunset API Server")
		fmt.Println("")
		fmt.Println("Usage: me-cli-sunset")
		fmt.Println("")
		fmt.Println("Environment variables:")
		fmt.Println("  SERVER_HOST        - HTTP server bind address (default: 0.0.0.0)")
		fmt.Println("  SERVER_PORT        - HTTP server port (default: 8080)")
		fmt.Println("  DATABASE_PATH      - SQLite database path (default: me_cli.db)")
		fmt.Println("  BASE_API_URL       - Base API URL (required)")
		fmt.Println("  BASE_CIAM_URL      - Base CIAM URL (required)")
		fmt.Println("  BASIC_AUTH         - Basic auth credentials")
		fmt.Println("  AX_FP_KEY          - Device fingerprint key")
		fmt.Println("  UA                 - User-Agent string")
		fmt.Println("  API_KEY            - API key")
		fmt.Println("  ENCRYPTED_FIELD_KEY - Encrypted field key")
		fmt.Println("  XDATA_KEY          - XData encryption key")
		fmt.Println("  AX_API_SIG_KEY     - AX API signature key")
		fmt.Println("  X_API_BASE_SECRET  - X API base secret")
		fmt.Println("  CIRCLE_MSISDN_KEY  - Circle MSISDN encryption key")
		os.Exit(0)
	}
}
