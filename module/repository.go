package module

import (
	history "vvinamp/package/history/repository"
	song "vvinamp/package/song/repository"
	user "vvinamp/package/user/repository"

	"go.uber.org/fx"
)

var RepositoryModule = fx.Provide(
	user.NewSQLRepository,
	history.NewSQLRepository,
	song.NewSQLRepository,
)
