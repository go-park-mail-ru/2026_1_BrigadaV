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

func TestReviewRepo_Create(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewReviewRepo(mockPool)

	review := &models.Review{
		UserID:    1,
		PlaceID:   1,
		Title:     testutil.PtrString("Great place"),
		Rating:    5,
		Comment:   "Excellent!",
		VisitDate: nil,
	}

	rows := mockPool.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(uint64(1), time.Now(), time.Now())

	mockPool.ExpectQuery(`INSERT INTO review \(user_id, place_id, title, rating, comment, visit_date\)`).
		WithArgs(review.UserID, review.PlaceID, review.Title, review.Rating, review.Comment, review.VisitDate).
		WillReturnRows(rows)

	err = repo.Create(context.Background(), review)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), review.ID)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestReviewRepo_GetByID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewReviewRepo(mockPool)

	title := "Great"
	rows := mockPool.NewRows([]string{"id", "user_id", "place_id", "title", "rating", "comment", "visit_date", "created_at", "updated_at"}).
		AddRow(uint64(1), uint64(1), uint64(1), &title, 5, "Excellent", nil, time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, user_id, place_id, title, rating, comment, visit_date, created_at, updated_at FROM review WHERE id = \$1`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	review, err := repo.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, uint64(1), review.ID)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestReviewRepo_GetByPlaceIDWithAuthor(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewReviewRepo(mockPool)

	title := "Great"
	avatar := "/avatar.jpg"

	rows := mockPool.NewRows([]string{"id", "title", "rating", "comment", "created_at", "user_id", "nickname", "avatar_url"}).
		AddRow(uint64(1), &title, 5, "Excellent", time.Now(), uint64(1), "johnny", &avatar)

	mockPool.ExpectQuery(`SELECT r\.id, r\.title, r\.rating, r\.comment, r\.created_at, u\.id, u\.nickname, u\.avatar_url`).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	reviews, err := repo.GetByPlaceIDWithAuthor(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, reviews, 1)
	assert.Equal(t, "johnny", reviews[0].Author.Nickname)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestReviewRepo_Delete(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	repo := NewReviewRepo(mockPool)

	mockPool.ExpectExec(`DELETE FROM review WHERE id = \$1`).
		WithArgs(uint64(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), 1)
	assert.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
