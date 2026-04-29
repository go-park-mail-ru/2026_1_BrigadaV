func (r *PlaceRepo) GetWithRatingAndLike(ctx context.Context, placeID, userID uint64) (*models.PlaceWithRating, error) {
	var place models.Place
	query := `SELECT id, name, description, photo_url, price, rating, review_count FROM place WHERE id = $1`
	err := r.db.QueryRow(ctx, query, placeID).Scan(
		&place.ID, &place.Name, &place.Description, &place.PhotoURL, &place.Price,
		&place.Rating, &place.ReviewCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get place: %w", err)
	}

	var isLiked bool
	if userID != 0 {
		_ = r.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM favorite WHERE user_id=$1 AND place_id=$2)`,
			userID, placeID,
		).Scan(&isLiked)
	}

	return &models.PlaceWithRating{
		ID:          place.ID,
		Name:        place.Name,
		Description: place.Description,
		PhotoURL:    place.PhotoURL,
		Price:       place.Price,
		Rating:      place.Rating,
		ReviewCount: int64(place.ReviewCount),
		IsLiked:     isLiked,
	}, nil
}