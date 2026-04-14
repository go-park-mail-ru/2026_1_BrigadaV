package repository

import (
	"context"
	"guidely-app/internal/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestReviewRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &ReviewRepo{db: nil}

	review := &models.Review{
		UserID:    1,
		PlaceID:   1,
		Title:     ptrString("Great place"),
		Rating:    5,
		Comment:   "Excellent!",
		VisitDate: nil,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO review (user_id, place_id, title, rating, comment, visit_date) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`)).
		WithArgs(review.UserID, review.PlaceID, review.Title, review.Rating, review.Comment, review.VisitDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(1, time.Now(), time.Now()))

	err = repo.Create(context.Background(), review)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), review.ID)
}

func TestReviewRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &ReviewRepo{db: nil}

	rows := sqlmock.NewRows([]string{"id", "user_id", "place_id", "title", "rating", "comment", "visit_date", "created_at", "updated_at"}).
		AddRow(1, 1, 1, "Great", 5, "Excellent", nil, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, place_id, title, rating, comment, visit_date, created_at, updated_at FROM review WHERE id = $1`)).
		WithArgs(1).
		WillReturnRows(rows)

	review, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, uint64(1), review.ID)
}

func TestReviewRepo_GetByPlaceIDWithAuthor(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &ReviewRepo{db: nil}

	rows := sqlmock.NewRows([]string{"id", "title", "rating", "comment", "created_at", "user_id", "nickname", "avatar_url"}).
		AddRow(1, "Great", 5, "Excellent", time.Now(), 1, "johnny", "/avatar.jpg")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT r.id, r.title, r.rating, r.comment, r.created_at, u.id, u.nickname, u.avatar_url FROM review r JOIN "user" u ON r.user_id = u.id WHERE r.place_id = $1 ORDER BY r.created_at DESC`)).
		WithArgs(1).
		WillReturnRows(rows)

	reviews, err := repo.GetByPlaceIDWithAuthor(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, reviews, 1)
	assert.Equal(t, "johnny", reviews[0].Author.Nickname)
}

func TestReviewRepo_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &ReviewRepo{db: nil}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM review WHERE id = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), 1)
	assert.NoError(t, err)
}

func ptrString(s string) *string {
	return &s
}
