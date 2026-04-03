package httpui

import (
	"go.uber.org/fx"

	"github.com/sonpham/vielish/server/internal/driving/httpui/handler"
	"github.com/sonpham/vielish/server/internal/driving/httpui/presenter"
)

var Module = fx.Module("httpui",
	fx.Provide(NewGin),
	fx.Provide(handler.NewHandler),
	fx.Provide(presenter.NewAuthPresenter),
	fx.Invoke(RegisterRoutes),
	fx.Invoke(RegisterLifecycle),
)
