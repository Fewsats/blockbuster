package store

import (
	"context"

	"github.com/fewsats/blockbuster/store/sqlc"
	"github.com/fewsats/blockbuster/video"
)

// Video methods
func (s *Store) CreateVideo(ctx context.Context, params video.CreateVideoParams) (*video.Video, error) {
	v, err := s.queries.CreateVideo(ctx, sqlc.CreateVideoParams{
		ExternalID:   params.ExternalID,
		UserID:       params.UserID,
		Title:        params.Title,
		Description:  params.Description,
		VideoUrl:     params.VideoURL,
		CoverUrl:     params.CoverURL,
		PriceInCents: params.PriceInCents,
	})

	if err != nil {
		return nil, err
	}

	return &video.Video{
		ID: v.ID,
	}, nil
}

func (s *Store) DeleteVideo(ctx context.Context, externalID string) error {
	return s.queries.DeleteVideo(ctx, externalID)
}

func (s *Store) GetVideoByExternalID(ctx context.Context, externalID string) (*video.Video, error) {
	v, err := s.queries.GetVideoByExternalID(ctx, externalID)
	if err != nil {
		return nil, err
	}

	return &video.Video{
		ID:           v.ID,
		ExternalID:   v.ExternalID,
		UserID:       v.UserID,
		Title:        v.Title,
		Description:  v.Description,
		VideoURL:     v.VideoUrl,
		CoverURL:     v.CoverUrl,
		PriceInCents: v.PriceInCents,
		TotalViews:   v.TotalViews,
		CreatedAt:    v.CreatedAt.Time,
	}, nil
}

func (s *Store) ListUserVideos(ctx context.Context,
	userID int64) ([]*video.Video, error) {

	videos, err := s.queries.ListUserVideos(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []*video.Video
	for _, v := range videos {
		result = append(result, &video.Video{
			ID:           v.ID,
			ExternalID:   v.ExternalID,
			UserID:       v.UserID,
			Title:        v.Title,
			Description:  v.Description,
			VideoURL:     v.VideoUrl,
			CoverURL:     v.CoverUrl,
			PriceInCents: v.PriceInCents,
			TotalViews:   v.TotalViews,
			CreatedAt:    v.CreatedAt.Time,
		})
	}

	return result, nil
}
