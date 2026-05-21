package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"guidely-app/internal/logger"
	"guidely-app/pkg/elasticsearch"
	"guidely-app/pkg/models"
	"strconv"

	"github.com/sirupsen/logrus"
)

type ElasticPlaceSearcher struct {
	client   *elasticsearch.Client
	placeRepo PlaceRepository
}

func NewElasticPlaceSearcher(client *elasticsearch.Client, placeRepo PlaceRepository) *ElasticPlaceSearcher {
	return &ElasticPlaceSearcher{
		client:    client,
		placeRepo: placeRepo,
	}
}

func (s *ElasticPlaceSearcher) Search(ctx context.Context, query string) ([]models.Place, error) {
	logger.Debug(ctx, "searching places via elasticsearch", logrus.Fields{"query": query})

	fields := []string{"name^3", "country^2", "locality^2", "description"}

	esQuery := elasticsearch.SearchRequest{
		Size: 50,
		Query: map[string]any{
			"bool": map[string]any{
				"should": []any{
					map[string]any{
						"multi_match": map[string]any{
							"query":  query,
							"fields": fields,
							"type":   "phrase_prefix",
							"boost":  3,
						},
					},
					map[string]any{
						"multi_match": map[string]any{
							"query":     query,
							"fields":    fields,
							"type":      "best_fields",
							"fuzziness": "AUTO",
							"boost":     1,
						},
					},
				},
				"minimum_should_match": 1,
			},
		},
	}

	resp, err := s.client.Search(ctx, elasticsearch.PlaceIndex, esQuery)
	if err != nil {
		logger.Error(ctx, "elasticsearch search failed", logrus.Fields{"error": err})
		return nil, fmt.Errorf("elasticsearch search: %w", err)
	}

	ids := make([]uint64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var doc elasticsearch.PlaceDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			logger.Error(ctx, "failed to unmarshal elasticsearch hit", logrus.Fields{"id": hit.ID, "error": err})
			continue
		}
		if doc.ID == 0 {
			parsed, err := strconv.ParseUint(hit.ID, 10, 64)
			if err != nil {
				continue
			}
			doc.ID = parsed
		}
		ids = append(ids, doc.ID)
	}

	if len(ids) == 0 {
		return []models.Place{}, nil
	}

	return s.placeRepo.GetByIDs(ctx, ids)
}
