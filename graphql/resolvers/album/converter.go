package album

// import (
// 	"database/sql"
// 	"time"

// 	"vvinamp/graphql/model"
// 	"vvinamp/internal/repository/album"
// )

// func ToGraphQL(dbAlbum *album.Album) *model.Album {
// 	if dbAlbum == nil {
// 		return nil
// 	}

// 	return &model.Album{
// 		ID:          dbAlbum.ID,
// 		Title:       dbAlbum.Title,
// 		ReleaseDate: formatDate(dbAlbum.ReleaseDate),
// 		CoverArt:    toStringPtr(dbAlbum.CoverArt),
// 		Artist:      convertArtist(dbAlbum),
// 		Genre:       convertGenre(dbAlbum),
// 	}
// }

// func toStringPtr(ns sql.NullString) *string {
// 	if !ns.Valid {
// 		return nil
// 	}
// 	return &ns.String
// }

// // Helper functions
// func formatDate(t time.Time) string {
// 	if t.IsZero() {
// 		return ""
// 	}
// 	return t.Format(time.RFC3339)
// }

// func convertArtist(a *album.Album) *model.Artist {
// 	return &model.Artist{
// 		ID:   a.ArtistID,
// 		Name: a.ArtistName,
// 	}
// }

// func convertGenre(a *album.Album) *model.Genre {
// 	if !a.GenreID.Valid || !a.GenreName.Valid {
// 		return nil
// 	}
// 	return &model.Genre{
// 		ID:   a.GenreID.String,
// 		Name: a.GenreName.String,
// 	}
// }
