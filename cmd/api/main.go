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

	"github.com/gorilla/csrf"
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
	placeService := service.NewPlaceService(placeRepo)
	profileService := service.NewProfileService(userRepo)
	tripService := service.NewTripService(tripRepo)
	reviewService := service.NewReviewService(reviewRepo)

	authHandler := handlers.NewAuthHandler(authService)
	placeHandler := handlers.NewPlaceHandler(placeService)
	profileHandler := handlers.NewProfileHandler(profileService)
	tripHandler := handlers.NewTripHandler(tripService)
	reviewHandler := handlers.NewReviewHandler(reviewService)

	authMiddleware := middleware.NewAuthMiddleware(sessionRepo)

	r := mux.NewRouter()

	r.HandleFunc("/api/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/api/places", placeHandler.List).Methods("GET")
	r.HandleFunc("/api/places/{id:[0-9]+}", placeHandler.Get).Methods("GET")
	r.HandleFunc("/api/reviews", reviewHandler.List).Methods("GET")

	r.HandleFunc("/api/logout", authMiddleware.Authenticate(authHandler.Logout)).Methods("POST")
	r.HandleFunc("/api/user/me", authMiddleware.Authenticate(authHandler.Me)).Methods("GET")
	r.HandleFunc("/api/profile", authMiddleware.Authenticate(profileHandler.GetProfile)).Methods("GET")
	r.HandleFunc("/api/profile", authMiddleware.Authenticate(profileHandler.UpdateProfile)).Methods("PUT") // объединённый

	r.HandleFunc("/api/trips", authMiddleware.Authenticate(tripHandler.List)).Methods("GET")
	r.HandleFunc("/api/trips", authMiddleware.Authenticate(tripHandler.Create)).Methods("POST")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.Get)).Methods("GET")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.Update)).Methods("PUT")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.Delete)).Methods("DELETE")

	r.HandleFunc("/api/reviews", authMiddleware.Authenticate(reviewHandler.Create)).Methods("POST")
	r.HandleFunc("/api/reviews/{id:[0-9]+}", authMiddleware.Authenticate(reviewHandler.Delete)).Methods("DELETE")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	csrfMiddleware := csrf.Protect(
		[]byte(cfg.JWTSecret),
		csrf.Secure(false),
		csrf.HttpOnly(true),
		csrf.Path("/"),
	)
	r.Use(csrfMiddleware)
	r.Use(middleware.CORS)

	log.Printf("Server started on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
