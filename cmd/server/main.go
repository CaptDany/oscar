package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/opencrm/opencrm/internal/api"
	"github.com/opencrm/opencrm/internal/api/handlers"
	"github.com/opencrm/opencrm/internal/api/middleware"
	"github.com/opencrm/opencrm/internal/config"
	"github.com/opencrm/opencrm/pkg/crypto"
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

	cryptoSvc := crypto.New(cfg.App.Secret)
	tokenManager := crypto.NewTokenManager(cfg.App.Secret)

	server := api.New()

	personHandler := &handlers.PersonHandler{}
	companyHandler := &handlers.CompanyHandler{}
	dealHandler := &handlers.DealHandler{}
	pipelineHandler := &handlers.PipelineHandler{}
	activityHandler := &handlers.ActivityHandler{}
	authHandler := &handlers.AuthHandler{}

	authMw := middleware.Auth(tokenManager)
	tenantMw := middleware.TenantResolver(nil)

	server.SetupRoutes(&handlers.Handlers{
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
