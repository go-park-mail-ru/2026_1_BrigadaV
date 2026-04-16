package main

import (
	_ "guidely-app/docs"
	"guidely-app/internal/config"
	"guidely-app/internal/db"
	"guidely-app/internal/handlers"
	"guidely-app/internal/middleware"
	"guidely-app/internal/repository"
	"guidely-app/internal/service"
	"log"
	"net/http"

	// "github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config load error:", err)
	}

	dbPool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("database connection error:", err)
	}
	defer dbPool.Close()

	userRepo := repository.NewUserRepo(dbPool)
	sessionRepo := repository.NewSessionRepo(dbPool)
	placeRepo := repository.NewPlaceRepo(dbPool)
	tripRepo := repository.NewTripRepo(dbPool)
	reviewRepo := repository.NewReviewRepo(dbPool)

	authService := service.NewAuthService(userRepo, sessionRepo)
	placeService := service.NewPlaceService(placeRepo, reviewRepo)
	profileService := service.NewProfileService(userRepo)
	tripService := service.NewTripService(tripRepo)
	reviewService := service.NewReviewService(reviewRepo)

	authHandler := handlers.NewAuthHandler(authService)
	placeHandler := handlers.NewPlaceHandler(placeService)
	profileHandler := handlers.NewProfileHandler(profileService)
	tripHandler := handlers.NewTripHandler(tripService)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	csrfHandler := handlers.NewCSRFHandler()

	authMiddleware := middleware.NewAuthMiddleware(sessionRepo)

	r := mux.NewRouter()

	r.HandleFunc("/api/register", authHandler.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/logout", authMiddleware.Authenticate(authHandler.Logout)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/user/me", authMiddleware.Authenticate(authHandler.Me)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/profile", authMiddleware.Authenticate(profileHandler.GetProfile)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/profile", authMiddleware.Authenticate(profileHandler.UpdateProfile)).Methods("PUT", "OPTIONS")

	r.HandleFunc("/api/places", placeHandler.List).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/places/{id:[0-9]+}", placeHandler.GetDetails).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/places/{id:[0-9]+}/reviews", placeHandler.GetReviews).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/trips", authMiddleware.Authenticate(tripHandler.List)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/trips", authMiddleware.Authenticate(tripHandler.Create)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.GetDetails)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.Update)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.Delete)).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}/places", authMiddleware.Authenticate(tripHandler.GetTripPlaces)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}/places", authMiddleware.Authenticate(tripHandler.AddPlace)).Methods("POST", "OPTIONS")
r.HandleFunc("/api/trips/{id:[0-9]+}/places/{placeId:[0-9]+}", authMiddleware.Authenticate(tripHandler.RemovePlace)).Methods("DELETE", "OPTIONS"


	r.HandleFunc("/api/reviews", authMiddleware.Authenticate(reviewHandler.Create)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/reviews/{id:[0-9]+}", authMiddleware.Authenticate(reviewHandler.Delete)).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/api/csrf-token", csrfHandler.GetToken).Methods("GET", "OPTIONS")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// csrfMiddleware := csrf.Protect(
	// 	[]byte(cfg.JWTSecret),
	// 	csrf.Secure(false),
	// 	csrf.HttpOnly(true),
	// 	csrf.Path("/"),
	// )
	// r.Use(csrfMiddleware)
	r.Use(middleware.CORS(cfg.FrontendURL))

	log.Printf("Server started on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
