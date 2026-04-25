package main

import (
	"os"
	_ "guidely-app/docs"
	"guidely-app/internal/config"
	"guidely-app/internal/db"
	"guidely-app/internal/handlers"
	"guidely-app/internal/logger"
	"guidely-app/internal/middleware"
	"guidely-app/internal/repository"
	"guidely-app/internal/service"
	"log"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	pb "guidely-app/pkg/pb/support/v1"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config load error:", err)
	}

	logger.Init("info")

	dbPool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("database connection error:", err)
	}
	defer dbPool.Close()

	dbAdapter := &repository.PgxPoolAdapter{Pool: dbPool}

	userRepo := repository.NewUserRepo(dbAdapter)
	sessionRepo := repository.NewSessionRepo(dbAdapter)
	placeRepo := repository.NewPlaceRepo(dbAdapter)
	tripRepo := repository.NewTripRepo(dbAdapter)
	reviewRepo := repository.NewReviewRepo(dbAdapter)

	authService := service.NewAuthService(userRepo, sessionRepo)
	placeService := service.NewPlaceService(placeRepo, reviewRepo)
	profileService := service.NewProfileService(userRepo)
	tripService := service.NewTripService(tripRepo)
	reviewService := service.NewReviewService(reviewRepo)

	authHandler := handlers.NewAuthHandler(authService)
	placeHandler := handlers.NewPlaceHandler(placeService, tripService)
	profileHandler := handlers.NewProfileHandler(profileService)
	tripHandler := handlers.NewTripHandler(tripService)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	csrfHandler := handlers.NewCSRFHandler()

	authMiddleware := middleware.NewAuthMiddleware(sessionRepo)

supportAddr := os.Getenv("SUPPORT_GRPC_ADDR")
if supportAddr == "" {
    supportAddr = "localhost:8084"
}

supportConn, err := grpc.Dial(supportAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: Failed to connect to support service: %v", err)
		log.Printf("Support features will be unavailable")
	} else {
		defer supportConn.Close()
		log.Printf("Connected to support gRPC service at %s", supportAddr)
	}
	
	var supportHandlers *handlers.SupportHandlers
	if supportConn != nil {
		supportClient := pb.NewSupportServiceClient(supportConn)
		supportHandlers = handlers.NewSupportHandlers(supportClient)
	}
	r := mux.NewRouter()
	
	r.Use(logger.Middleware)
	
	if supportHandlers != nil {
		logger.Log.Info("Support routes enabled")
		r.HandleFunc("/api/support/tickets", authMiddleware.Authenticate(supportHandlers.CreateTicket)).Methods("POST", "OPTIONS")
		r.HandleFunc("/api/support/tickets", authMiddleware.Authenticate(supportHandlers.ListMyTickets)).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/support/tickets/{id}", authMiddleware.Authenticate(supportHandlers.GetTicket)).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/support/tickets/{id}/messages", authMiddleware.Authenticate(supportHandlers.SendMessage)).Methods("POST", "OPTIONS")
		
		r.HandleFunc("/api/admin/support/stats", authMiddleware.Authenticate(supportHandlers.GetStats)).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/admin/support/tickets/open", authMiddleware.Authenticate(supportHandlers.ListOpenTickets)).Methods("GET", "OPTIONS")
		r.HandleFunc("/api/admin/support/tickets/{id}/reply", authMiddleware.Authenticate(supportHandlers.ReplyAsAdmin)).Methods("POST", "OPTIONS")
		r.HandleFunc("/api/admin/support/tickets/{id}/status", authMiddleware.Authenticate(supportHandlers.UpdateTicketStatus)).Methods("PATCH", "OPTIONS")
	}
	
	r.HandleFunc("/api/register", authHandler.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/logout", authMiddleware.Authenticate(authHandler.Logout)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/user/me", authMiddleware.Authenticate(authHandler.Me)).Methods("GET", "OPTIONS")
	
	r.HandleFunc("/api/profile", authMiddleware.Authenticate(profileHandler.GetProfile)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/profile", authMiddleware.Authenticate(profileHandler.UpdateProfile)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/profile/avatar", authMiddleware.Authenticate(profileHandler.UploadAvatar)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/profile/avatar", authMiddleware.Authenticate(profileHandler.GetAvatar)).Methods("GET", "OPTIONS")
	
	r.HandleFunc("/api/places", placeHandler.List).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/places/{id:[0-9]+}", placeHandler.GetDetails).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/places/{id:[0-9]+}/reviews", placeHandler.GetReviews).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/places/{id:[0-9]+}/in-trip", authMiddleware.Authenticate(placeHandler.CheckPlaceInTrip)).Methods("GET", "OPTIONS")
	
	r.HandleFunc("/api/trips", authMiddleware.Authenticate(tripHandler.List)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/trips", authMiddleware.Authenticate(tripHandler.Create)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.GetDetails)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.Update)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.Delete)).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}/places", authMiddleware.Authenticate(tripHandler.GetTripPlaces)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}/places", authMiddleware.Authenticate(tripHandler.AddPlace)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}/places/{placeId:[0-9]+}", authMiddleware.Authenticate(tripHandler.RemovePlace)).Methods("DELETE", "OPTIONS")
	
	r.HandleFunc("/api/reviews", authMiddleware.Authenticate(reviewHandler.Create)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/reviews/{id:[0-9]+}", authMiddleware.Authenticate(reviewHandler.Delete)).Methods("DELETE", "OPTIONS")


	r.HandleFunc("/api/csrf-token", csrfHandler.GetToken).Methods("GET", "OPTIONS")
	
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	
	csrfMiddleware := csrf.Protect(
			[]byte(cfg.JWTSecret),
			csrf.Secure(false),
			csrf.HttpOnly(true),
			csrf.Path("/"),
			csrf.TrustedOrigins([]string{"guidely.ru"}),
		)
		r.Use(csrfMiddleware)
		r.Use(middleware.CORS(cfg.FrontendURL))
		
	logger.Log.Info("Server started on :" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
