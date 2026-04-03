package presenter

import (
	"github.com/gin-gonic/gin"

	appcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	"github.com/sonpham/vielish/server/pkg/httpbase"
)

type AuthPresenter struct{}

func NewAuthPresenter() *AuthPresenter { return &AuthPresenter{} }

func (p *AuthPresenter) Tokens(c *gin.Context, status int, out *appcore.TokenOutput) {
	httpbase.Success(c, status, gin.H{
		"access_token":  out.AccessToken,
		"refresh_token": out.RefreshToken,
		"expires_in":    out.ExpiresIn,
	})
}
