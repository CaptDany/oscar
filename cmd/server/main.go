package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/api"
	"github.com/oscar/oscar/internal/api/handlers"
	"github.com/oscar/oscar/internal/api/middleware"
	"github.com/oscar/oscar/internal/config"
	"github.com/oscar/oscar/internal/db/repositories"
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

	cryptoSvc := crypto.New()
	tokenManager := crypto.NewTokenManager(cfg.App.Secret)

	personRepo := repositories.NewPersonRepository(pool)
	companyRepo := repositories.NewCompanyRepository(pool)
	dealRepo := repositories.NewDealRepository(pool)
	pipelineRepo := repositories.NewPipelineRepository(pool)
	activityRepo := repositories.NewActivityRepository(pool)
	activityAssocRepo := repositories.NewActivityAssociationRepository(pool)
	tenantRepo := repositories.NewTenantRepository(pool)
	userRepo := repositories.NewUserRepository(pool)

	server := api.New()

	personHandler := handlers.NewPersonHandler(personRepo)
	companyHandler := handlers.NewCompanyHandler(companyRepo)
	dealHandler := handlers.NewDealHandler(dealRepo, pipelineRepo)
	pipelineHandler := handlers.NewPipelineHandler(pipelineRepo)
	activityHandler := handlers.NewActivityHandler(activityRepo, activityAssocRepo)
	roleRepo := repositories.NewRoleRepository(pool)
	authHandler := handlers.NewAuthHandler(userRepo, tenantRepo, roleRepo, cryptoSvc, tokenManager)

	authMw := middleware.Auth(tokenManager)
	tenantMw := middleware.TenantResolver(tenantRepo)

	server.SetupRoutes(&api.Handlers{
		Auth:     authHandler,
		Person:   personHandler,
		Company:  companyHandler,
		Deal:     dealHandler,
		Pipeline: pipelineHandler,
		Activity: activityHandler,
	}, authMw, tenantMw)

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

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
