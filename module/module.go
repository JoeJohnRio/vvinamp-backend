package module

import (
	"vvinamp/config"
	"vvinamp/controller"
	"vvinamp/database/mysql"
	"vvinamp/logging"
	"vvinamp/server"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		config.New,
		logging.New,
		server.New,
		// Database
		mysql.New,
		// Controller
		controller.NewGraphQLController,
		controller.NewAuth,
	),
	RepositoryModule,
	ServiceModule,
)
