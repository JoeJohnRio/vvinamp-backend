package user

import (
	"spotify-clone/graphql/model"
	"spotify-clone/internal/repository/user"
	"spotify-clone/internal/utils"
)

func ToGraphQLQuickPick(p user.QuickPick) *model.QuickPick {
	return &model.QuickPick{
		TrackID:    p.TrackID,
		Title:      p.Title,
		ArtistName: p.ArtistName,
		PlayCount:  p.PlayCount,
		CoverArt:   utils.ToStringPtr(p.CoverArt),
	}
}

func ToGraphQLQuickPicks(picks []user.QuickPick) []*model.QuickPick {
	var gqlPicks []*model.QuickPick
	for _, p := range picks {
		gqlPicks = append(gqlPicks, ToGraphQLQuickPick(p))
	}
	return gqlPicks
}
