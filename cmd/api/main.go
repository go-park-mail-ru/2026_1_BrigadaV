package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "guidely-app/docs"
	authrepo "guidely-app/internal/auth/repository"
	"guidely-app/internal/handlers"
	"guidely-app/internal/logger"
	"guidely-app/internal/middleware"
	"guidely-app/internal/repository"
	"guidely-app/internal/service"
	"guidely-app/pkg/config"
	"guidely-app/pkg/db"
	"guidely-app/pkg/elasticsearch"
	"guidely-app/pkg/metrics"
	"guidely-app/pkg/storage"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
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

	s3Client, err := storage.NewS3Client(cfg)
	if err != nil {
		log.Printf("S3 init warning: %v; continuing without S3", err)
		s3Client = nil
	}
	if s3Client == nil {
		log.Println("S3 client is nil – avatar will be stored locally")
	}

	// gRPC подключения
	authConn, err := grpc.NewClient(getEnv("AUTH_GRPC_ADDR", "localhost:8085"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	authClient := pbauth.NewAuthServiceClient(authConn)

	albumConn, err := grpc.NewClient(getEnv("ALBUM_GRPC_ADDR", "localhost:8086"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to album service: %v", err)
	}
	albumClient := pbalbum.NewAlbumServiceClient(albumConn)

	reviewConn, err := grpc.NewClient(getEnv("REVIEW_GRPC_ADDR", "localhost:8087"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to review service: %v", err)
	}
	reviewClient := pbreview.NewReviewServiceClient(reviewConn)

	dbAdapter := &repository.PgxPoolAdapter{Pool: dbPool}
	authAdapter := &authrepo.PgxPoolAdapter{Pool: dbPool}

	// Репозитории
	placeRepo := repository.NewPlaceRepo(dbAdapter)
	tripRepo := repository.NewTripRepo(dbAdapter)
	categoryRepo := repository.NewCategoryRepo(dbAdapter)
	countryRepo := repository.NewCountryRepo(dbAdapter)
	reviewRepo := repository.NewReviewRepo(dbAdapter)
	userRepo := authrepo.NewUserRepo(authAdapter)
	sessionRepo := authrepo.NewSessionRepo(authAdapter)
	tripMemberRepo := repository.NewTripMemberRepo(dbAdapter)
	tripInviteRepo := repository.NewTripInviteRepo(dbAdapter)

	esClient := elasticsearch.NewClient(cfg.ElasticSearchURL)
	placeIndexer := elasticsearch.NewPlaceIndexer(esClient)

	go func() {
		if err := waitForElastic(context.Background(), esClient, 30, 5*time.Second); err != nil {
			log.Printf("warning: elasticsearch not ready: %v", err)
			return
		}
		if err := placeIndexer.EnsureIndex(context.Background()); err != nil {
			log.Printf("warning: failed to ensure elasticsearch index: %v", err)
			return
		}
		if err := indexAllPlaces(context.Background(), placeRepo, placeIndexer); err != nil {
			log.Printf("warning: failed to index places on startup: %v", err)
		}
	}()

	elasticSearcher := repository.NewElasticPlaceSearcher(esClient, placeRepo)

	placeService := service.NewPlaceService(placeRepo, reviewRepo, elasticSearcher)
	tripService := service.NewTripService(tripRepo, tripMemberRepo, tripInviteRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	countryService := service.NewCountryService(countryRepo)
	profileService := service.NewProfileService(userRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authClient, cfg)
	albumHandler := handlers.NewAlbumHandler(albumClient, s3Client)
	reviewHandler := handlers.NewReviewHandler(reviewClient)
	placeHandler := handlers.NewPlaceHandler(placeService, tripService)
	profileHandler := handlers.NewProfileHandler(profileService, s3Client)
	tripHandler := handlers.NewTripHandler(tripService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	countryHandler := handlers.NewCountryHandler(countryService)
	csrfHandler := handlers.NewCSRFHandler()
	yandexHandler := handlers.NewYandexOAuthHandler(
		cfg.YandexClientID,
		cfg.YandexClientSecret,
		cfg.YandexRedirectURL,
		cfg.FrontendURL,
		cfg.SecureCookies,
		userRepo,
		sessionRepo,
	)

	authMiddleware := middleware.NewAuthMiddleware(sessionRepo)

	csrfMiddleware := csrf.Protect(
		[]byte(cfg.CSRFSecret),
		csrf.Secure(cfg.SecureCookies),
		csrf.Path("/"),
		csrf.TrustedOrigins(cfg.AllowedOrigins),
	)

	r := mux.NewRouter()
	r.Use(logger.Middleware)
	r.Use(middleware.CORS(cfg.AllowedOrigins...))
	r.Use(metrics.HTTPMetricsMiddleware)

	csrfTokenRouter := r.PathPrefix("/api").Subrouter()
	csrfTokenRouter.Use(csrfMiddleware)
	csrfTokenRouter.HandleFunc("/csrf-token", csrfHandler.GetToken).Methods("GET", "OPTIONS")

	// Публичные эндпоинты
	public := r.PathPrefix("/api").Subrouter()
	public.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	public.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")
	public.HandleFunc("/csrf-token", csrfHandler.GetToken).Methods("GET", "OPTIONS")
	public.HandleFunc("/share/view/{token}", tripHandler.ViewSharedTrip).Methods("GET")
	public.HandleFunc("/share/edit/{token}", tripHandler.AcceptInviteRedirect).Methods("GET")

	public.HandleFunc("/places", placeHandler.List).Methods("GET", "OPTIONS")
	public.HandleFunc("/places/search", placeHandler.Search).Methods("GET", "OPTIONS")
	public.HandleFunc("/places/{id:[0-9]+}", placeHandler.GetDetails).Methods("GET", "OPTIONS")
	public.HandleFunc("/places/{id:[0-9]+}/reviews", placeHandler.GetReviews).Methods("GET", "OPTIONS")
	public.HandleFunc("/categories", categoryHandler.List).Methods("GET", "OPTIONS")
	public.HandleFunc("/categories/{id:[0-9]+}", categoryHandler.Get).Methods("GET", "OPTIONS")
	public.HandleFunc("/countries", countryHandler.List).Methods("GET", "OPTIONS")
	public.HandleFunc("/countries/{id:[0-9]+}/localities", countryHandler.GetWithLocalities).Methods("GET", "OPTIONS")
	public.HandleFunc("/auth/yandex/login", yandexHandler.Login).Methods("GET", "OPTIONS")
	public.HandleFunc("/auth/yandex/callback", yandexHandler.Callback).Methods("GET", "OPTIONS")

	// Защищённые без CSRF
	authOnly := r.PathPrefix("/api").Subrouter()
	authOnly.Use(authMiddleware.Authenticate)

	authOnly.HandleFunc("/share/edit/{token}", tripHandler.AcceptInviteAPI).Methods("POST", "OPTIONS")
	authOnly.HandleFunc("/share/view/{token}", tripHandler.JoinViewShareAPI).Methods("POST", "OPTIONS")

	authOnly.HandleFunc("/profile/avatar", profileHandler.GetAvatar).Methods("GET", "OPTIONS")
	authOnly.HandleFunc("/profile/avatar", profileHandler.UploadAvatar).Methods("POST", "OPTIONS")
	authOnly.HandleFunc("/profile", profileHandler.UpdateProfile).Methods("PUT", "OPTIONS")
	authOnly.HandleFunc("/albums/{id:[0-9]+}/photos", albumHandler.AddPhoto).Methods("POST", "OPTIONS")
	authOnly.HandleFunc("/logout", authHandler.Logout).Methods("POST", "OPTIONS")
	authOnly.HandleFunc("/reviews", reviewHandler.Create).Methods("POST", "OPTIONS")
	authOnly.HandleFunc("/reviews/{id:[0-9]+}", reviewHandler.Delete).Methods("DELETE", "OPTIONS")
	authOnly.HandleFunc("/trips/{id:[0-9]+}/places", tripHandler.AddPlace).Methods("POST", "OPTIONS")
	authOnly.HandleFunc("/trips/{id:[0-9]+}", tripHandler.Update).Methods("PUT", "OPTIONS")
	authOnly.HandleFunc("/trips/{id:[0-9]+}", tripHandler.Delete).Methods("DELETE", "OPTIONS")
	authOnly.HandleFunc("/places/{id:[0-9]+}/in-trip", placeHandler.CheckPlaceInTrip).Methods("GET", "OPTIONS")
	authOnly.HandleFunc("/trips/{id:[0-9]+}/share/view", tripHandler.CreateViewShareLink).Methods("POST", "OPTIONS")
	authOnly.HandleFunc("/trips/{id:[0-9]+}/share/edit", tripHandler.CreateEditShareLink).Methods("POST", "OPTIONS")
	authOnly.HandleFunc("/trips/{id:[0-9]+}/members", tripHandler.GetTripMembers).Methods("GET", "OPTIONS")
	authOnly.HandleFunc("/trips/{id:[0-9]+}/members/{member_id:[0-9]+}", tripHandler.RemoveMember).Methods("DELETE", "OPTIONS")
	authOnly.HandleFunc("/trips", tripHandler.Create).Methods("POST", "OPTIONS")

	// Защищённые с CSRF
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(authMiddleware.Authenticate)
	protected.Use(csrfMiddleware)

	protected.HandleFunc("/user/me", authHandler.Me).Methods("GET", "OPTIONS")
	protected.HandleFunc("/profile", profileHandler.GetProfile).Methods("GET", "OPTIONS")
	protected.HandleFunc("/trips", tripHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/trips/{id:[0-9]+}", tripHandler.GetDetails).Methods("GET", "OPTIONS")
	protected.HandleFunc("/trips/{id:[0-9]+}/places", tripHandler.GetTripPlaces).Methods("GET", "OPTIONS")
	protected.HandleFunc("/trips/{id:[0-9]+}/recommended-places", tripHandler.GetRecommendedPlaces).Methods("GET", "OPTIONS")
	protected.HandleFunc("/trips/{id:[0-9]+}/places/{placeId:[0-9]+}", tripHandler.RemovePlace).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/trips/{tripID:[0-9]+}/album", albumHandler.GetByTrip).Methods("GET", "OPTIONS")
	protected.HandleFunc("/albums/{id:[0-9]+}/photos", albumHandler.GetPhotos).Methods("GET", "OPTIONS")
	protected.HandleFunc("/albums/{id:[0-9]+}/photos/{photoId:[0-9]+}", albumHandler.RemovePhoto).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/categories", categoryHandler.Create).Methods("POST", "OPTIONS")
	protected.HandleFunc("/categories/{id:[0-9]+}", categoryHandler.Update).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/categories/{id:[0-9]+}", categoryHandler.Delete).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/trips/{id:[0-9]+}/export/pdf", tripHandler.ExportTripToPDF).Methods("GET", "OPTIONS")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	logger.Log.Info("Server started on :" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func indexAllPlaces(ctx context.Context, repo repository.PlaceRepository, indexer *elasticsearch.PlaceIndexer) error {
	places, err := repo.GetAll(ctx, repository.PlaceFilter{})
	if err != nil {
		return err
	}
	for _, p := range places {
		if err := indexer.IndexPlace(ctx, p); err != nil {
			log.Printf("warning: failed to index place %d: %v", p.ID, err)
		}
	}
	log.Printf("indexed %d places in elasticsearch", len(places))
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func waitForElastic(ctx context.Context, client *elasticsearch.Client, attempts int, delay time.Duration) error {
	for i := 0; i < attempts; i++ {
		_, status, err := client.HealthCheck(ctx)
		if err == nil && status < 400 {
			return nil
		}
		log.Printf("elasticsearch not ready (attempt %d/%d), retrying in %s...", i+1, attempts, delay)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("elasticsearch not ready after %d attempts", attempts)
}
