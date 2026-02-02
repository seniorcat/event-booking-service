package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	pb "laschool.ru/event-booking-service/api/v1/booking"
	_ "laschool.ru/event-booking-service/docs"
	"laschool.ru/event-booking-service/internal/booking"
	"laschool.ru/event-booking-service/internal/cache"
	"laschool.ru/event-booking-service/internal/config"
	"laschool.ru/event-booking-service/internal/db"
	"laschool.ru/event-booking-service/internal/event"
	grpcHandlers "laschool.ru/event-booking-service/internal/grpc/handlers"
	httprouter "laschool.ru/event-booking-service/internal/http"
	"laschool.ru/event-booking-service/internal/http/middleware"
	di "laschool.ru/event-booking-service/pkg/container"
)

// @title           Event Booking Service API
// @version         1.0
// @description     Сервис для управления событиями и бронированиями.
// @BasePath        /
// @schemes         http
// @host            localhost:8080
// @securityDefinitions.apikey  Bearer
// @in header
// @name Authorization
func main() {
	fmt.Println("Booking Service started...")

	// загружаем конфиг
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize DI container
	ctn, err := di.Instance(nil, nil)
	if err != nil {
		log.Fatalf("failed to initialize DI container: %v", err)
	}

	// Optional migrations
	if cfg.Database.AutoMigrate {
		database := ctn.Get(db.DIDatabase).(*sqlx.DB)
		if err := db.RunMigrations(context.Background(), database, "deploy/migrations"); err != nil {
			log.Fatalf("migrations failed: %v", err)
		}
	}

	// маршруты
	mux := httprouter.NewRouter()
	// логирование сервера
	loggingMux := middleware.LoggingMiddleware(mux)
	muxWithLogAndPanic := middleware.PanicMiddleware(loggingMux)

	// старт сервера
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      muxWithLogAndPanic,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("HTTP server listening on port %d...", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}()

	//gRPC server

	grpcServer := grpc.NewServer()
	bookingService := ctn.Get(booking.DIBookingService).(booking.Service)
	eventService := ctn.Get(event.DIEventService).(event.Service)
	cacheService := ctn.Get(cache.DICacheService).(cache.Service)

	bookingHandler := grpcHandlers.NewGRPCBookingServer(bookingService, eventService, cacheService)
	pb.RegisterBookingServiceServer(grpcServer, bookingHandler)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}

	go func() {
		log.Println("gRPC server listening on :50051...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, stopping servers...")

	// Stop accepting new gRPC connections and wait for ongoing
	grpcServer.GracefulStop()

	// Shutdown HTTP with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Servers stopped")

}
