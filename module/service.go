package module

import (
	user "vvinamp/package/user/service"

	"go.uber.org/fx"
)

// ServiceModule is Repositories fx module
var ServiceModule = fx.Provide(
	user.NewService,
)
