package appcore

import "go.uber.org/fx"

var Module = fx.Module("vocab-appcore",
	fx.Provide(NewUseCase),
)
