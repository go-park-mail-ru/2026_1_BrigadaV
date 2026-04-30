package main

import (
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

	// "github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	authConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	authClient := pbauth.NewAuthServiceClient(authConn)

	albumConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to album service: %v", err)
	}
	albumClient := pbalbum.NewAlbumServiceClient(albumConn)

	placeRepo := repository.NewPlaceRepo(dbAdapter)
	reviewRepo := repository.NewReviewRepo(dbAdapter)
	tripRepo := repository.NewTripRepo(dbAdapter)
	categoryRepo := repository.NewCategoryRepo(dbPool)
	userRepo := repository.NewUserRepo(dbAdapter)
	sessionRepo := repository.NewSessionRepo(dbAdapter)

	placeService := service.NewPlaceService(placeRepo, reviewRepo)
	tripService := service.NewTripService(tripRepo)
	reviewService := service.NewReviewService(reviewRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	profileService := service.NewProfileService(userRepo)

	authHandler := handlers.NewAuthHandler(authClient)
	albumHandler := handlers.NewAlbumHandler(albumClient)
	placeHandler := handlers.NewPlaceHandler(placeService, tripService)
	profileHandler := handlers.NewProfileHandler(profileService)
	tripHandler := handlers.NewTripHandler(tripService)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	csrfHandler := handlers.NewCSRFHandler()

	authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)

	r := mux.NewRouter()
	r.Use(logger.Middleware)
	r.Use(middleware.CORS(cfg.FrontendURL))

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

	r.HandleFunc("/api/albums/{id:[0-9]+}", authMiddleware.Authenticate(albumHandler.Update)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/albums/{id:[0-9]+}", authMiddleware.Authenticate(albumHandler.Delete)).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/api/categories", categoryHandler.List).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/categories/{id:[0-9]+}", categoryHandler.Get).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/categories", authMiddleware.Authenticate(categoryHandler.Create)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/categories/{id:[0-9]+}", authMiddleware.Authenticate(categoryHandler.Update)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/categories/{id:[0-9]+}", authMiddleware.Authenticate(categoryHandler.Delete)).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/api/reviews", authMiddleware.Authenticate(reviewHandler.Create)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/reviews/{id:[0-9]+}", authMiddleware.Authenticate(reviewHandler.Delete)).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/api/csrf-token", csrfHandler.GetToken).Methods("GET", "OPTIONS")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// CSRF middleware (закомментирован)
	// csrfMiddleware := csrf.Protect(
	// 	[]byte(cfg.JWTSecret),
	// 	csrf.Secure(false),
	// 	csrf.HttpOnly(true),
	// 	csrf.Path("/"),
	// )
	// r.Use(csrfMiddleware)
	r.Use(middleware.CORS(cfg.FrontendURL))

	logger.Log.Info("Server started on :" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
