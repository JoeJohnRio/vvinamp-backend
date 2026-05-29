package module

import (
	"vvinamp/config"
	"vvinamp/controller"
	"vvinamp/controller/auth"
	"vvinamp/database/postgre"
	"vvinamp/database/redis"
	"vvinamp/logging"
	"vvinamp/package/streaming"
	"vvinamp/server"

	"go.uber.org/fx"
)

// 1. Define a tiny struct to handle the tagging
type HlsOut struct {
	fx.Out
	Converter streaming.HLSConverter `name:"hlsConverter"`
}

var Module = fx.Options(
	fx.Provide(
		config.New,
		logging.New,
		server.New,
		postgre.New,
		redis.New,
		controller.NewGraphQLController,
		controller.NewHomeController,
		controller.NewUploadController,

		// 2. Replace the simple constructor with this wrapper
		func(target controller.FileServerControllerTarget) (controller.Result, HlsOut) {
			// A. Create the controller as usual
			res := controller.NewFileServerController(target)

			// B. Extract the concrete pointer (assuming res.Controller is the interface or pointer)
			// We check if it implements the HLSConverter interface
			var converter streaming.HLSConverter

			// Note: Depending on how your 'Result' struct is defined, you might need a type assertion.
			// If res.Controller is already *FileServerController, this works directly:
			if c, ok := res.Controller.(streaming.HLSConverter); ok {
				converter = c
			} else {
				// Fallback if res.Controller is the concrete type directly
				// (This part depends on your exact Result struct definition, but usually it's one of these)
				converter = res.Controller.(streaming.HLSConverter)
			}

			// C. Return BOTH:
			//    - 'res' connects it to the Controller Group (so routes work)
			//    - 'HlsOut' connects it to Streaming Service with the NAME tag "hlsConverter"
			return res, HlsOut{
				Converter: converter,
			}
		},

		auth.NewAuth,
	),
	RepositoryModule,
	ServiceModule,
)
