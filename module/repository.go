package module

import (
	artist "vvinamp/package/artist/repository"
	history "vvinamp/package/history/repository"
	playlist "vvinamp/package/playlist/repository"
	search "vvinamp/package/search/repository"
	song "vvinamp/package/song/repository"
	user "vvinamp/package/user/repository"

	"go.uber.org/fx"
)

var RepositoryModule = fx.Provide(
	artist.NewSQLRepository,
	user.NewSQLRepository,
	history.NewSQLRepository,
	song.NewSQLRepository,
	search.NewSQLRepository,
	playlist.NewSQLRepository,
)
