package main

import (
	"go.uber.org/fx"

	authAppcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	authDomain "github.com/sonpham/vielish/server/internal/domain/auth"
	authDriven "github.com/sonpham/vielish/server/internal/driven/auth"
	"github.com/sonpham/vielish/server/internal/driven/database"
	"github.com/sonpham/vielish/server/internal/driving/httpui"
	"github.com/sonpham/vielish/server/pkg/config"
	pkglog "github.com/sonpham/vielish/server/pkg/log"
	"github.com/sonpham/vielish/server/pkg/tracing"
)

func main() {
	fx.New(
		// Infrastructure
		config.Module,
		pkglog.Module,
		tracing.Module,
		database.Module,

		// Auth feature
		fx.Module("auth",
			authDomain.Module,
			authAppcore.Module,
			authDriven.Module,
		),

		// HTTP server
		httpui.Module,
	).Run()
}
