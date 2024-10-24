package store

import (
	"context"
	"database/sql"

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
		CoverUrl:     params.CoverURL,
		PriceInCents: params.PriceInCents,
		CreatedAt:    s.clock.Now(),
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

func (s *Store) GetVideoByExternalID(ctx context.Context,
	externalID string) (*video.Video, error) {

	v, err := s.queries.GetVideoByExternalID(ctx, externalID)
	if err != nil {
		return nil, err
	}

	return &video.Video{
		ID:                v.ID,
		ExternalID:        v.ExternalID,
		UserID:            v.UserID,
		Title:             v.Title,
		Description:       v.Description,
		CoverURL:          v.CoverUrl,
		PriceInCents:      v.PriceInCents,
		TotalPurchases:    v.TotalPurchases,
		TotalViews:        v.TotalViews,
		ThumbnailURL:      v.ThumbnailUrl.String,
		DurationInSeconds: v.DurationInSeconds.Float64,
		SizeInBytes:       v.SizeInBytes.Int64,
		InputHeight:       int32(v.InputHeight.Int64),
		InputWidth:        int32(v.InputWidth.Int64),
		ReadyToStream:     v.ReadyToStream,
		CreatedAt:         v.CreatedAt,
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
			ID:         v.ID,
			ExternalID: v.ExternalID,
			UserID:     v.UserID,

			Title:             v.Title,
			Description:       v.Description,
			CoverURL:          v.CoverUrl,
			PriceInCents:      v.PriceInCents,
			TotalViews:        v.TotalViews,
			TotalPurchases:    v.TotalPurchases,
			ThumbnailURL:      v.ThumbnailUrl.String,
			HlsURL:            v.HlsUrl.String,
			DashURL:           v.DashUrl.String,
			DurationInSeconds: v.DurationInSeconds.Float64,
			SizeInBytes:       v.SizeInBytes.Int64,
			InputHeight:       int32(v.InputHeight.Int64),
			InputWidth:        int32(v.InputWidth.Int64),
			ReadyToStream:     v.ReadyToStream,

			CreatedAt: v.CreatedAt,
		})
	}

	return result, nil
}

func (s *Store) UpdateCloudflareInfo(ctx context.Context, externalID string,
	params *video.CloudflareVideoInfo) (*video.Video, error) {

	v, err := s.queries.UpdateCloudflareInfo(ctx, sqlc.UpdateCloudflareInfoParams{
		ExternalID: externalID,
		ThumbnailUrl: sql.NullString{
			String: params.ThumbnailURL,
			Valid:  params.ThumbnailURL != "",
		},
		DurationInSeconds: sql.NullFloat64{
			Float64: params.DurationInSeconds,
			Valid:   params.DurationInSeconds != 0,
		},
		SizeInBytes: sql.NullInt64{
			Int64: params.SizeInBytes,
			Valid: params.SizeInBytes != 0,
		},
		InputHeight: sql.NullInt64{
			Int64: int64(params.InputHeight),
			Valid: params.InputHeight != 0,
		},
		InputWidth: sql.NullInt64{
			Int64: int64(params.InputWidth),
			Valid: params.InputWidth != 0,
		},
		DashUrl: sql.NullString{
			String: params.DashURL,
			Valid:  params.DashURL != "",
		},
		HlsUrl: sql.NullString{
			String: params.HLSURL,
			Valid:  params.HLSURL != "",
		},
		ReadyToStream: params.ReadyToStream,
	})
	if err != nil {
		return nil, err
	}
	return &video.Video{
		ID:                v.ID,
		ExternalID:        v.ExternalID,
		UserID:            v.UserID,
		Title:             v.Title,
		Description:       v.Description,
		CoverURL:          v.CoverUrl,
		HlsURL:            v.HlsUrl.String,
		DashURL:           v.DashUrl.String,
		PriceInCents:      v.PriceInCents,
		TotalViews:        v.TotalViews,
		ThumbnailURL:      v.ThumbnailUrl.String,
		DurationInSeconds: v.DurationInSeconds.Float64,
		SizeInBytes:       v.SizeInBytes.Int64,
		InputHeight:       int32(v.InputHeight.Int64),
		InputWidth:        int32(v.InputWidth.Int64),
		ReadyToStream:     v.ReadyToStream,
		CreatedAt:         v.CreatedAt,
	}, nil
}

func (s *Store) IncrementVideoViews(ctx context.Context, externalID string) error {
	return s.queries.IncrementVideoViews(ctx, externalID)
}

func (s *Store) UpdateVideoInfo(ctx context.Context, externalID string, params *video.UpdateVideoInfoParams) (*video.Video, error) {
	v, err := s.queries.UpdateVideoInfo(ctx, sqlc.UpdateVideoInfoParams{
		ExternalID: externalID,
		Title: sql.NullString{
			String: params.Title,
			Valid:  params.Title != "",
		},
		Description: sql.NullString{
			String: params.Description,
			Valid:  params.Description != "",
		},
		PriceInCents: sql.NullInt64{
			Int64: params.PriceInCents,
			Valid: params.PriceInCents != 0,
		},
	})
	if err != nil {
		return nil, err
	}
	return &video.Video{
		ID:           v.ID,
		ExternalID:   v.ExternalID,
		UserID:       v.UserID,
		Title:        v.Title,
		Description:  v.Description,
		CoverURL:     v.CoverUrl,
		PriceInCents: v.PriceInCents,
		TotalViews:   v.TotalViews,

		ThumbnailURL:      v.ThumbnailUrl.String,
		HlsURL:            v.HlsUrl.String,
		DashURL:           v.DashUrl.String,
		DurationInSeconds: v.DurationInSeconds.Float64,
		SizeInBytes:       v.SizeInBytes.Int64,
		InputHeight:       int32(v.InputHeight.Int64),
		InputWidth:        int32(v.InputWidth.Int64),
		ReadyToStream:     v.ReadyToStream,
		CreatedAt:         v.CreatedAt,
	}, nil
}
