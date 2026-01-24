package module

import (
	"vvinamp/config"
	"vvinamp/controller"
	"vvinamp/controller/auth"
	"vvinamp/database/postgre"
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
		postgre.New,
		// Controller
		controller.NewGraphQLController,
		controller.NewHomeController,
		controller.NewFileServerController,
		auth.NewAuth,
	),
	RepositoryModule,
	ServiceModule,
)
