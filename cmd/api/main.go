package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "guidely-app/docs"
	authrepo "guidely-app/internal/auth/repository"
	"guidely-app/internal/handlers"
	"guidely-app/internal/logger"
	"guidely-app/internal/middleware"
	"guidely-app/internal/repository"
	"guidely-app/internal/service"
	"guidely-app/pkg/config"
	"guidely-app/pkg/db"
	"guidely-app/pkg/storage"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbalbum "guidely-app/pkg/pb/album"
	pbauth "guidely-app/pkg/pb/auth"
	pbreview "guidely-app/pkg/pb/review"
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

	// ИНИЦИАЛИЗАЦИЯ S3 С ВОЗМОЖНОСТЬЮ ПРОДОЛЖИТЬ БЕЗ НЕГО
	s3Client, err := storage.NewS3Client(cfg)
	if err != nil {
		// Ошибка уже залогирована внутри NewS3Client, но на всякий случай
		log.Printf("S3 init failed: %v; continuing without S3 features", err)
		s3Client = nil
	}
	if s3Client == nil {
		log.Println("S3 client is nil – avatar upload will not work")
	}

	authConn, err := grpc.Dial(getEnv("AUTH_GRPC_ADDR", "localhost:8085"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	authClient := pbauth.NewAuthServiceClient(authConn)

	albumConn, err := grpc.Dial(getEnv("ALBUM_GRPC_ADDR", "localhost:8086"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to album service: %v", err)
	}
	albumClient := pbalbum.NewAlbumServiceClient(albumConn)

	reviewConn, err := grpc.Dial(getEnv("REVIEW_GRPC_ADDR", "localhost:8087"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to review service: %v", err)
	}
	reviewClient := pbreview.NewReviewServiceClient(reviewConn)

	dbAdapter := &repository.PgxPoolAdapter{Pool: dbPool}
	authAdapter := &authrepo.PgxPoolAdapter{Pool: dbPool}

	placeRepo := repository.NewPlaceRepo(dbAdapter)
	tripRepo := repository.NewTripRepo(dbAdapter)
	categoryRepo := repository.NewCategoryRepo(dbAdapter)
	reviewRepo := repository.NewReviewRepo(dbAdapter)
	userRepo := authrepo.NewUserRepo(authAdapter)
	sessionRepo := authrepo.NewSessionRepo(authAdapter)

	placeService := service.NewPlaceService(placeRepo, reviewRepo)
	tripService := service.NewTripService(tripRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	profileService := service.NewProfileService(userRepo)

	authHandler := handlers.NewAuthHandler(authClient)
	albumHandler := handlers.NewAlbumHandler(albumClient)
	reviewHandler := handlers.NewReviewHandler(reviewClient)
	placeHandler := handlers.NewPlaceHandler(placeService, tripService)
	profileHandler := handlers.NewProfileHandler(profileService, s3Client)
	tripHandler := handlers.NewTripHandler(tripService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	csrfHandler := handlers.NewCSRFHandler()

	authMiddleware := middleware.NewAuthMiddleware(sessionRepo)

	r := mux.NewRouter()
	r.Use(logger.Middleware)
	r.Use(middleware.CORS(cfg.AllowedOrigins...))

	r.HandleFunc("/api/register", authHandler.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/logout", authMiddleware.Authenticate(authHandler.Logout)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/user/me", authMiddleware.Authenticate(authHandler.Me)).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/profile", authMiddleware.Authenticate(profileHandler.GetProfile)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/profile", authMiddleware.Authenticate(profileHandler.UpdateProfile)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/profile/avatar", authMiddleware.Authenticate(profileHandler.UploadAvatar)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/profile/avatar", authMiddleware.Authenticate(profileHandler.GetAvatar)).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/places", placeHandler.List).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/places/search", placeHandler.Search).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/places/{id:[0-9]+}", placeHandler.GetDetails).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/places/{id:[0-9]+}/reviews", placeHandler.GetReviews).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/places/{id:[0-9]+}/in-trip", authMiddleware.Authenticate(placeHandler.CheckPlaceInTrip)).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/reviews", authMiddleware.Authenticate(reviewHandler.Create)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/reviews/{id:[0-9]+}", authMiddleware.Authenticate(reviewHandler.Delete)).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/api/trips", authMiddleware.Authenticate(tripHandler.List)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/trips", authMiddleware.Authenticate(tripHandler.Create)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.GetDetails)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.Update)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}", authMiddleware.Authenticate(tripHandler.Delete)).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}/places", authMiddleware.Authenticate(tripHandler.GetTripPlaces)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}/places", authMiddleware.Authenticate(tripHandler.AddPlace)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/trips/{id:[0-9]+}/places/{placeId:[0-9]+}", authMiddleware.Authenticate(tripHandler.RemovePlace)).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/api/trips/{tripID:[0-9]+}/album", authMiddleware.Authenticate(albumHandler.GetByTrip)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/albums/{id:[0-9]+}/photos", authMiddleware.Authenticate(albumHandler.AddPhoto)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/albums/{id:[0-9]+}/photos/{photoId:[0-9]+}", authMiddleware.Authenticate(albumHandler.RemovePhoto)).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/api/albums/{id:[0-9]+}/photos", authMiddleware.Authenticate(albumHandler.GetPhotos)).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/categories", categoryHandler.List).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/categories/{id:[0-9]+}", categoryHandler.Get).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/categories", authMiddleware.Authenticate(categoryHandler.Create)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/categories/{id:[0-9]+}", authMiddleware.Authenticate(categoryHandler.Update)).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/categories/{id:[0-9]+}", authMiddleware.Authenticate(categoryHandler.Delete)).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/api/csrf-token", csrfHandler.GetToken).Methods("GET", "OPTIONS")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	csrfMiddleware := csrf.Protect(
		[]byte(cfg.CSRFSecret),
		csrf.Secure(false),
		csrf.Path("/"),
		csrf.TrustedOrigins(cfg.AllowedOrigins),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Warn(r.Context(), "CSRF token invalid", logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "CSRF token invalid or missing"})
		})),
	)

	handler := csrfMiddleware(r)

	logger.Log.Info("Server started on :" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
