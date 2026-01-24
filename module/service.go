package module

import (
	artist "vvinamp/package/artist/service"
	history "vvinamp/package/history/service"
	song "vvinamp/package/song/service"
	user "vvinamp/package/user/service"

	"go.uber.org/fx"
)

// ServiceModule is Repositories fx module
var ServiceModule = fx.Provide(
	artist.NewService,
	user.NewService,
	history.NewService,
	song.NewService,
)
