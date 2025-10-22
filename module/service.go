package module

import (
	history "vvinamp/package/history/service"
	user "vvinamp/package/user/service"

	"go.uber.org/fx"
)

// ServiceModule is Repositories fx module
var ServiceModule = fx.Provide(
	user.NewService,
	history.NewService,
)
