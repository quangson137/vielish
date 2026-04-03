package domain

import "go.uber.org/fx"

var Module = fx.Module("vocab-domain",
	fx.Provide(NewService),
)
