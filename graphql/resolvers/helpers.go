package resolvers

import (
	"encoding/base64"
	"fmt"
	"time"

	"vvinamp/graphql/model"
	youtubeService "vvinamp/internal/services/youtube"
	dbModel "vvinamp/model"
	"vvinamp/package/search"
)

// convertToSearchItems converts search results to GraphQL models
func convertToSearchItems(
	tracks []*search.TrackWithDetails,
	albums []*search.AlbumWithArtist,
	artists []*dbModel.Artist,
	playlists []*dbModel.Playlist,
	podcastEpisodes []*dbModel.PodcastEpisode,
	podcastSeries []*dbModel.PodcastSeries,
) []model.SearchItem {
	var items []model.SearchItem

	for _, t := range tracks {
		items = append(items, &model.Track{
			ID:            t.ID.String(),
			Title:         t.Title,
			Duration:      t.Duration,
			StreamCount:   t.StreamCount,
			AudioFileURL:  t.AudioFileURL,
			AlbumCoverURL: t.AlbumCoverURL,
			ArtistName:    t.ArtistName,
			Explicit:      t.Explicit,
			Lyrics: func() *string {
				if t.Lyrics == "" {
					return nil
				}
				return &t.Lyrics
			}(),
			ReleaseDate: t.ReleaseDate,
		})
	}

	for _, a := range albums {
		items = append(items, &model.Album{
			ID:          a.ID.String(),
			Title:       a.Title,
			ReleaseDate: a.ReleaseDate,
			CoverArt: func() *string {
				if a.CoverArt == "" {
					return nil
				}
				return &a.CoverArt
			}(),
		})
	}

	for _, a := range artists {
		items = append(items, &model.Artist{
			ID:               a.ID,
			Name:             a.Name,
			Bio:              a.Bio,
			ProfileImage:     a.ProfileImage,
			Verified:         a.Verified,
			MonthlyListeners: int(a.MonthlyListeners),
		})
	}

	for _, p := range playlists {
		items = append(items, &model.Playlist{
			ID:         p.ID,
			Title:      p.Title,
			IsPublic:   p.IsPublic,
			CoverImage: p.CoverImage,
		})
	}

	for _, e := range podcastEpisodes {
		items = append(items, &model.PodcastEpisode{
			ID:            e.ID,
			Title:         e.Title,
			EpisodeNumber: e.EpisodeNumber,
			SeasonNumber:  e.SeasonNumber,
			Description:   e.Description,
			ReleaseDate:   formatTime(e.ReleaseDate),
			Duration:      e.Duration,
			AudioFileURL:  e.AudioFileURL,
			CoverArt:      e.CoverArt,
		})
	}

	for _, s := range podcastSeries {
		items = append(items, &model.PodcastSeries{
			ID:          s.ID,
			Title:       s.Title,
			Description: s.Description,
			CoverArt:    s.CoverArt,
			Publisher:   s.Publisher,
		})
	}

	return items
}

func formatTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func encodeCursor(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("cursor:%d", offset)))
}

func decodeCursor(cursor string) (int, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	var offset int
	_, err = fmt.Sscanf(string(decoded), "cursor:%d", &offset)
	return offset, err
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func ptr[T any](v T) *T { return &v }

func convertDownloadTask(task *youtubeService.DownloadTask) *model.DownloadTask {
	return &model.DownloadTask{
		ID:         task.ID,
		Title:      strPtr(task.Title),
		YoutubeURL: task.YoutubeURL,
		Status:     task.Status,
		Progress:   task.Progress,
		FilePath:   strPtr(task.FilePath),
		Error:      strPtr(task.Error),
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
	}
}
