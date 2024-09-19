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
		UserEmail:    params.UserEmail,
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

// func (s *Store) GetVideo(ctx context.Context, id int64) (*sqlc.Video, error) {
// 	return s.queries.GetVideo(ctx, id)
// }

// func (s *Store) ListVideos(ctx context.Context, limit, offset int32) ([]*sqlc.Video, error) {
// 	return s.queries.ListVideos(ctx, sqlc.ListVideosParams{
// 		Limit:  limit,
// 		Offset: offset,
// 	})
// }
