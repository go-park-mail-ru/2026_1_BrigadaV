package repository

import (
	"context"
	"testing"
	"time"

	"guidely-app/internal/models"
	"guidely-app/internal/testutil"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserRepo_Create(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewUserRepo(mockPool)

	user := &models.User{
		Login:        "test@example.com",
		Nickname:     "tester",
		AvatarURL:    "/avatar.jpg",
		PasswordHash: "hash123",
		Country:      testutil.PtrString("USA"),
		City:         testutil.PtrString("NYC"),
		About:        testutil.PtrString("About me"),
	}

	rows := mockPool.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(uint64(1), time.Now(), time.Now())

	mockPool.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(user.Login, user.Nickname, user.AvatarURL, user.PasswordHash, user.Country, user.City, user.About).
		WillReturnRows(rows)

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), user.ID)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepo_GetByLogin_Found(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewUserRepo(mockPool)

	country := "USA"
	city := "NYC"
	about := "About me"

	rows := mockPool.NewRows([]string{"id", "login", "nickname", "avatar_url", "password_hash", "country", "city", "about", "has_reviews", "created_at", "updated_at"}).
		AddRow(uint64(1), "test@example.com", "tester", "/avatar.jpg", "hash123", &country, &city, &about, false, time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at FROM "user" WHERE login = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	user, err := repo.GetByLogin(context.Background(), "test@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Login)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepo_GetByLogin_NotFound(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewUserRepo(mockPool)

	mockPool.ExpectQuery(`SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at FROM "user" WHERE login = \$1`).
		WithArgs("notfound@example.com").
		WillReturnRows(mockPool.NewRows([]string{"id", "login", "nickname", "avatar_url", "password_hash", "country", "city", "about", "has_reviews", "created_at", "updated_at"}))

	user, err := repo.GetByLogin(context.Background(), "notfound@example.com")
	assert.NoError(t, err)
	assert.Nil(t, user)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepo_GetByNickname_Found(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewUserRepo(mockPool)

	country := "USA"
	city := "NYC"
	about := "About me"

	rows := mockPool.NewRows([]string{"id", "login", "nickname", "avatar_url", "password_hash", "country", "city", "about", "has_reviews", "created_at", "updated_at"}).
		AddRow(uint64(1), "test@example.com", "tester", "/avatar.jpg", "hash123", &country, &city, &about, false, time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at FROM "user" WHERE nickname = \$1`).
		WithArgs("tester").
		WillReturnRows(rows)

	user, err := repo.GetByNickname(context.Background(), "tester")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "tester", user.Nickname)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepo_GetByID_Found(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewUserRepo(mockPool)

	country := "USA"
	city := "NYC"
	about := "About me"

	rows := mockPool.NewRows([]string{"id", "login", "nickname", "avatar_url", "password_hash", "country", "city", "about", "has_reviews", "created_at", "updated_at"}).
		AddRow(uint64(1), "test@example.com", "tester", "/avatar.jpg", "hash123", &country, &city, &about, false, time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, login, nickname, avatar_url, password_hash, country, city, about, has_reviews, created_at, updated_at FROM "user" WHERE id = \$1`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	user, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, uint64(1), user.ID)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepo_Update(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewUserRepo(mockPool)

	user := &models.User{
		ID:         1,
		Login:      "updated@example.com",
		Nickname:   "updated",
		AvatarURL:  "/new-avatar.jpg",
		Country:    testutil.PtrString("Canada"),
		City:       testutil.PtrString("Toronto"),
		About:      testutil.PtrString("Updated about"),
		HasReviews: true,
	}

	rows := mockPool.NewRows([]string{"updated_at"}).AddRow(time.Now())

	mockPool.ExpectQuery(`UPDATE "user" SET login = \$1, nickname = \$2, avatar_url = \$3, country = \$4, city = \$5, about = \$6, has_reviews = \$7, updated_at = NOW\(\) WHERE id = \$8 RETURNING updated_at`).
		WithArgs(user.Login, user.Nickname, user.AvatarURL, user.Country, user.City, user.About, user.HasReviews, uint64(1)).
		WillReturnRows(rows)

	err = repo.Update(context.Background(), user)
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
