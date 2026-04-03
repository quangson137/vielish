package domain

import "go.uber.org/fx"

var Module = fx.Module("auth-domain",
	fx.Provide(NewService),
)
