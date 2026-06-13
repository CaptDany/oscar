package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/oscar/oscar/internal/api"
	"github.com/oscar/oscar/internal/api/handlers"
	"github.com/oscar/oscar/internal/api/middleware"
	"github.com/oscar/oscar/internal/cache"
	"github.com/oscar/oscar/internal/config"
	"github.com/oscar/oscar/internal/db/repositories"
	"github.com/oscar/oscar/internal/email"
	"github.com/oscar/oscar/internal/storage"
	"github.com/oscar/oscar/pkg/crypto"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	var redisClient *redis.Client
	var rateLimiter *middleware.InMemoryRateLimiter
	var cacheSvc cache.Cache

	if cfg.Redis.URL != "" {
		u, err := url.Parse(cfg.Redis.URL)
		if err != nil {
			log.Printf("[Cache] Invalid Redis URL: %v", err)
		} else {
			var password string
			if u.User != nil {
				password, _ = u.User.Password()
			}

			redisClient = redis.NewClient(&redis.Options{
				Addr:        u.Host,
				Password:    password,
				DB:          cfg.Redis.DB,
				PoolSize:    cfg.Redis.PoolSize,
				PoolTimeout: cfg.Redis.PoolTimeout,
			})

			if err := redisClient.Ping(context.Background()).Err(); err != nil {
				log.Printf("[Cache] Redis unavailable, using in-memory fallback: %v", err)
				redisClient = nil
			} else {
				log.Println("[Cache] Redis connected successfully")
				cacheSvc = cache.NewRedisCache(redisClient)
			}
		}
	}

	rateLimiter = middleware.NewRateLimiter(redisClient, 100, time.Minute)

	cryptoSvc := crypto.New()
	tokenManager := crypto.NewTokenManager(cfg.App.Secret)

	emailClient := email.NewEmailClient(&cfg.Email)

	personRepo := repositories.NewPersonRepository(pool)
	companyRepo := repositories.NewCompanyRepository(pool)
	customFieldRepo := repositories.NewCustomFieldRepository(pool)
	dealRepo := repositories.NewDealRepository(pool)
	pipelineRepo := repositories.NewPipelineRepository(pool)
	activityRepo := repositories.NewActivityRepository(pool)
	activityAssocRepo := repositories.NewActivityAssociationRepository(pool)
	tenantRepo := repositories.NewTenantRepository(pool)
	userRepo := repositories.NewUserRepository(pool)
	notificationRepo := repositories.NewNotificationRepository(pool)
	teamRepo := repositories.NewTeamRepository(pool)
	productRepo := repositories.NewProductRepository(pool)
	lineItemRepo := repositories.NewLineItemRepository(pool)
	brandingRepo := repositories.NewBrandingRepository(pool)
	invitationRepo := repositories.NewInvitationRepository(pool)
	auditLogRepo := repositories.NewAuditLogRepository(pool)
	apiKeyRepo := repositories.NewAPIKeyRepository(pool)
	attachmentRepo := repositories.NewAttachmentRepository(pool)

	if cacheSvc != nil {
		tenantRepo.SetCache(cacheSvc)
		brandingRepo.SetCache(cacheSvc)
	}

	server := api.New()

	personHandler := handlers.NewPersonHandler(personRepo)
	companyHandler := handlers.NewCompanyHandler(companyRepo)
	dealHandler := handlers.NewDealHandler(dealRepo, pipelineRepo)
	pipelineHandler := handlers.NewPipelineHandler(pipelineRepo)
	activityHandler := handlers.NewActivityHandler(activityRepo, activityAssocRepo)
	roleRepo := repositories.NewRoleRepository(pool)
	authHandler := handlers.NewAuthHandlerWithInvitations(userRepo, tenantRepo, roleRepo, invitationRepo, cryptoSvc, tokenManager, emailClient, cfg.App.BaseURL, cfg.App.FrontendURL)
	oauthHandler := handlers.NewOAuthHandler(userRepo, tenantRepo, roleRepo, cryptoSvc, tokenManager, nil, cfg.App.BaseURL, &cfg.OAuth)

	if cfg.OAuth.GoogleClientID == "" {
		log.Println("[OAuth] Google OAuth is not configured (OAUTH_GOOGLE_CLIENT_ID not set)")
	}
	if cfg.OAuth.AppleClientID == "" {
		log.Println("[OAuth] Apple OAuth is not configured (OAUTH_APPLE_CLIENT_ID not set)")
	}

	r2Client, err := storage.NewR2Client(&cfg.R2)
	if err != nil {
		log.Printf("Warning: Failed to create R2 client: %v (upload features disabled)", err)
		r2Client = nil
	} else {
		if err := r2Client.EnsureBucket(context.Background()); err != nil {
			log.Printf("Warning: Failed to ensure R2 bucket: %v (upload features disabled)", err)
		}
	}

	var userHandler *handlers.UserHandler
	var uploadHandler *handlers.UploadHandler
	var attachmentHandler *handlers.AttachmentHandler
	if r2Client != nil {
		userHandler = handlers.NewUserHandler(userRepo, roleRepo, personRepo, dealRepo, r2Client)
		uploadHandler = handlers.NewUploadHandler(r2Client, userRepo)
		attachmentHandler = handlers.NewAttachmentHandler(attachmentRepo, r2Client)
	}

	authMw := middleware.Auth(tokenManager)
	tenantPool := repositories.NewTenantPool(pool)
	tenantMw := middleware.TenantResolver(tenantRepo, tenantPool)

	notificationHandler := handlers.NewNotificationHandler(notificationRepo)
	teamHandler := handlers.NewTeamHandler(teamRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	settingsHandler := handlers.NewSettingsHandler(tenantRepo, brandingRepo)
	invitationHandler := handlers.NewInvitationHandler(invitationRepo, userRepo, roleRepo, tenantRepo, cryptoSvc, &handlers.MockEmailSender{})
	customFieldHandler := handlers.NewCustomFieldHandler(customFieldRepo)
	lineItemHandler := handlers.NewLineItemHandler(lineItemRepo, dealRepo)
	auditLogHandler := handlers.NewAuditLogHandler(auditLogRepo)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyRepo)

	server.SetupRoutes(&api.Handlers{
		Auth:         authHandler,
		OAuth:        oauthHandler,
		Person:       personHandler,
		Company:      companyHandler,
		Deal:         dealHandler,
		Pipeline:     pipelineHandler,
		Activity:     activityHandler,
		User:         userHandler,
		Upload:       uploadHandler,
		Notification: notificationHandler,
		Team:         teamHandler,
		Product:      productHandler,
		Settings:     settingsHandler,
		Invitation:   invitationHandler,
		CustomField:  customFieldHandler,
		LineItem:     lineItemHandler,
		AuditLog:     auditLogHandler,
		APIKey:       apiKeyHandler,
		Attachment:   attachmentHandler,
	}, authMw, tenantMw, rateLimiter)

	addr := fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port)

	go func() {
		log.Printf("Starting server on %s", addr)
		if err := server.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	if redisClient != nil {
		_ = redisClient.Close()
	}

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
