package driven

import (
	"go.uber.org/fx"

	domain "github.com/sonpham/vielish/server/internal/domain/auth"
)

var Module = fx.Module("auth-driven",
	fx.Provide(
		fx.Annotate(
			NewRepository,
			fx.As(new(domain.Repository)),
		),
	),
)
