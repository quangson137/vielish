package appcore

import "go.uber.org/fx"

var Module = fx.Module("auth-appcore",
	fx.Provide(NewUseCase),
)
