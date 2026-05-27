package repository

import (
	"context"
	"guidely-app/pkg/models"
)

type CountryRepo struct {
	db DB
}

func NewCountryRepo(db DB) *CountryRepo {
	return &CountryRepo{db: db}
}

func (r *CountryRepo) GetAll(ctx context.Context) ([]models.Country, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, created_at FROM country ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var countries []models.Country
	for rows.Next() {
		var c models.Country
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		countries = append(countries, c)
	}
	return countries, nil
}

func (r *CountryRepo) GetLocalitiesByCountryID(ctx context.Context, countryID uint64) ([]models.Locality, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, latitude, longitude FROM locality WHERE country_id = $1 ORDER BY name`,
		countryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var localities []models.Locality
	for rows.Next() {
		var l models.Locality
		if err := rows.Scan(&l.ID, &l.Name, &l.Latitude, &l.Longitude); err != nil {
			return nil, err
		}
		localities = append(localities, l)
	}
	return localities, nil
}
