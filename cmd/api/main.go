package main

import (
	_ "guidely-app/docs"
	"guidely-app/internal/config"
	"guidely-app/internal/db"
	"guidely-app/internal/handlers"
	"guidely-app/internal/middleware"
	"guidely-app/internal/models"
	"guidely-app/internal/repository"
	"guidely-app/internal/service"
	"guidely-app/internal/storage"
	"guidely-app/internal/utils"
	"log"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config load error:", err)
	}

	store := storage.NewMemoryStore()
	initTestData(store)

	userRepo := repository.NewUserRepo(store)
	sessionRepo := repository.NewSessionRepo(store)
	placeRepo := repository.NewPlaceRepo(store)

	authService := service.NewAuthService(userRepo, sessionRepo)
	placeService := service.NewPlaceService(placeRepo)
	profileService := service.NewProfileService(userRepo)

	authHandler := handlers.NewAuthHandler(authService)
	placeHandler := handlers.NewPlaceHandler(placeService)
	profileHandler := handlers.NewProfileHandler(profileService)

	authMiddleware := middleware.NewAuthMiddleware(sessionRepo)

	dbPool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("database connection error:", err)
	}
	defer dbPool.Close()

	tripRepo := repository.NewTripRepo(dbPool)
	reviewRepo := repository.NewReviewRepo(dbPool)

	tripService := service.NewTripService(tripRepo)
	reviewService := service.NewReviewService(reviewRepo)

	tripHandler := handlers.NewTripHandler(tripService)
	reviewHandler := handlers.NewReviewHandler(reviewService)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/register", middleware.CORS(authHandler.Register))
	mux.HandleFunc("/api/login", middleware.CORS(authHandler.Login))
	mux.HandleFunc("/api/", middleware.CORS(placeHandler.List))
	mux.HandleFunc("/api/places/", middleware.CORS(placeHandler.Get))
	mux.HandleFunc("/api/reviews", middleware.CORS(reviewHandler.List))

	mux.HandleFunc("/api/logout", middleware.CORS(authMiddleware.Authenticate(authHandler.Logout)))
	mux.HandleFunc("/api/user/me", middleware.CORS(authMiddleware.Authenticate(authHandler.Me)))
	mux.HandleFunc("/api/profile", middleware.CORS(authMiddleware.Authenticate(profileHandler.GetProfile)))
	mux.HandleFunc("/api/profile/update", middleware.CORS(authMiddleware.Authenticate(profileHandler.UpdateProfile)))
	mux.HandleFunc("/api/profile/change-password", middleware.CORS(authMiddleware.Authenticate(profileHandler.ChangePassword)))

	mux.HandleFunc("/api/trips", middleware.CORS(authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tripHandler.List(w, r)
		} else if r.Method == http.MethodPost {
			tripHandler.Create(w, r)
		} else {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})))
	mux.HandleFunc("/api/trips/", middleware.CORS(authMiddleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tripHandler.Get(w, r)
		case http.MethodPut:
			tripHandler.Update(w, r)
		case http.MethodDelete:
			tripHandler.Delete(w, r)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})))

	mux.HandleFunc("/api/reviews/", middleware.CORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authMiddleware.Authenticate(reviewHandler.Create)(w, r)
		} else if r.Method == http.MethodDelete {
			authMiddleware.Authenticate(reviewHandler.Delete)(w, r)
		} else {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Printf("Server started on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}

func initTestData(store *storage.MemoryStore) {
	catHotel := models.Category{ID: 1, Name: "Hotel", Description: "Hotels and accommodations"}
	catMuseum := models.Category{ID: 2, Name: "Museum", Description: "Art and historical museums"}
	catHistorical := models.Category{ID: 3, Name: "Historical Site", Description: "Ancient ruins and landmarks"}
	catSquare := models.Category{ID: 4, Name: "Square", Description: "Public squares and plazas"}
	catResort := models.Category{ID: 5, Name: "Resort", Description: "Resorts and retreats"}

	locGramado := models.Locality{ID: 1, Name: "Грамаду", Country: "Бразилия", Latitude: ptr(-29.3733), Longitude: ptr(-50.8762)}
	locParis := models.Locality{ID: 2, Name: "Париж", Country: "Франция", Latitude: ptr(48.8566), Longitude: ptr(2.3522)}
	locRome := models.Locality{ID: 3, Name: "Рим", Country: "Италия", Latitude: ptr(41.9028), Longitude: ptr(12.4964)}
	locBarcelona := models.Locality{ID: 4, Name: "Барселона", Country: "Испания", Latitude: ptr(41.3851), Longitude: ptr(2.1734)}
	locAmsterdam := models.Locality{ID: 5, Name: "Амстердам", Country: "Нидерланды", Latitude: ptr(52.3676), Longitude: ptr(4.9041)}
	locBali := models.Locality{ID: 6, Name: "Бали", Country: "Индонезия", Latitude: ptr(-8.4095), Longitude: ptr(115.1889)}

	now := time.Now()

	store.Places[1] = models.Place{
		ID:          1,
		Name:        "Hotel Estalagem St Hubertus",
		Description: "Charming hotel in Gramado",
		Price:       2370000,
		Locality:    locGramado,
		Category:    catHotel,
		Photos: []models.PlacePhoto{
			{ID: 1, PlaceID: 1, FilePath: "mock/place/rcmd1.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.Places[2] = models.Place{
		ID:          2,
		Name:        "Hotel Ritta Höppner",
		Description: "Cozy hotel in Gramado",
		Price:       1138100,
		Locality:    locGramado,
		Category:    catHotel,
		Photos: []models.PlacePhoto{
			{ID: 2, PlaceID: 2, FilePath: "mock/place/rcmd2.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.Places[3] = models.Place{
		ID:          3,
		Name:        "Rodin Musée",
		Description: "Museum dedicated to Auguste Rodin",
		Price:       126900,
		Locality:    locParis,
		Category:    catMuseum,
		Photos: []models.PlacePhoto{
			{ID: 3, PlaceID: 3, FilePath: "mock/place/rcmd3.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.Places[4] = models.Place{
		ID:          4,
		Name:        "Roman Forum",
		Description: "Ancient Roman forum",
		Price:       126900,
		Locality:    locRome,
		Category:    catHistorical,
		Photos: []models.PlacePhoto{
			{ID: 4, PlaceID: 4, FilePath: "mock/place/rcmd4.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.Places[5] = models.Place{
		ID:          5,
		Name:        "Basílica de Santa María del Pi",
		Description: "Gothic church in Barcelona",
		Price:       199400,
		Locality:    locBarcelona,
		Category:    catHistorical,
		Photos: []models.PlacePhoto{
			{ID: 5, PlaceID: 5, FilePath: "mock/place/rcmd5.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.Places[6] = models.Place{
		ID:          6,
		Name:        "De Hallen Amsterdam",
		Description: "Cultural complex in Amsterdam",
		Price:       3398800,
		Locality:    locAmsterdam,
		Category:    catMuseum,
		Photos: []models.PlacePhoto{
			{ID: 6, PlaceID: 6, FilePath: "mock/place/rcmd6.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.Places[7] = models.Place{
		ID:          7,
		Name:        "Amnaya Resort Kuta",
		Description: "Resort in Bali",
		Price:       584400,
		Locality:    locBali,
		Category:    catResort,
		Photos: []models.PlacePhoto{
			{ID: 7, PlaceID: 7, FilePath: "mock/place/rcmd7.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.Places[8] = models.Place{
		ID:          8,
		Name:        "Plaça Reial",
		Description: "Historic square in Barcelona",
		Price:       1236900,
		Locality:    locBarcelona,
		Category:    catSquare,
		Photos: []models.PlacePhoto{
			{ID: 8, PlaceID: 8, FilePath: "mock/place/rcmd8.png", IsMain: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	hashed, _ := utils.HashPassword("123456")
	john := models.User{
		ID:           store.NextUserID,
		Login:        "john@example.com",
		Nickname:     "johnny",
		AvatarURL:    "mock/user-avatar/john.jpg",
		PasswordHash: hashed,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	store.Users[john.ID] = john
	store.UsersByEmail[john.Login] = john.ID
	store.UsersByNickname[john.Nickname] = john.ID
	store.NextUserID++

	store.UserLikes[1] = make(map[uint64]bool)
}

func ptr(f float64) *float64 {
	return &f
}
