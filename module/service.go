package module

import (
	artist "vvinamp/package/artist/service"
	history "vvinamp/package/history/service"
	playlist "vvinamp/package/playlist/service"
	search "vvinamp/package/search/service"
	song "vvinamp/package/song/service"
	user "vvinamp/package/user/service"
	streamingService "vvinamp/package/streaming"

	"go.uber.org/fx"
)

// ServiceModule is Repositories fx module
var ServiceModule = fx.Provide(
	artist.NewService,
	user.NewService,
	history.NewService,
	song.NewService,
	search.NewService,
	streamingService.NewService,
	playlist.NewService,
)
