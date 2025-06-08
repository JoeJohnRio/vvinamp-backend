package module

import (
	user "vvinamp/package/user/repository"

	"go.uber.org/fx"
)

var RepositoryModule = fx.Provide(
	user.NewSQLRepository,
)
