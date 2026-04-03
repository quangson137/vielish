package driven

import (
	"go.uber.org/fx"

	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
)

var Module = fx.Module("vocab-driven",
	fx.Provide(
		fx.Annotate(
			NewRepository,
			fx.As(new(domain.Repository)),
		),
	),
)
