package genre

// import (
// 	"vvinamp/graphql/model"
// 	"vvinamp/internal/repository/genre"
// 	"vvinamp/internal/utils"
// )

// func ToGraphQL(dbGenres []genre.Genre) []*model.Genre {
// 	if dbGenres == nil {
// 		return nil
// 	}

// 	var gqlGenres []*model.Genre
// 	for i := range dbGenres {
// 		// take pointer of each value
// 		dbGenre := &dbGenres[i]
// 		gqlGenres = append(gqlGenres, ToGraphQLGenre(dbGenre))
// 	}

// 	return gqlGenres
// }

// func ToGraphQLGenre(dbGenre *genre.Genre) *model.Genre {
// 	if dbGenre == nil {
// 		return nil
// 	}

// 	return &model.Genre{
// 		ID:          dbGenre.ID,
// 		Name:        dbGenre.Name,
// 		Description: utils.ToStringPtr(dbGenre.Description),
// 	}
// }
