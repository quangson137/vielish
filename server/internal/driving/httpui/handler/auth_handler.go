package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	appcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	domain "github.com/sonpham/vielish/server/internal/domain/auth"
	"github.com/sonpham/vielish/server/internal/driving/httpui/presenter"
	"github.com/sonpham/vielish/server/pkg/httpbase"
)

// UseCaseInterface allows the handler to be tested with stubs.
type UseCaseInterface interface {
	Register(ctx context.Context, in appcore.RegisterInput) (*appcore.TokenOutput, error)
	Login(ctx context.Context, in appcore.LoginInput) (*appcore.TokenOutput, error)
	Refresh(ctx context.Context, token string) (*appcore.TokenOutput, error)
	Logout(ctx context.Context, token string) error
}

type registerRequest struct {
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type Handler struct {
	useCase   UseCaseInterface
	presenter *presenter.AuthPresenter
}

func NewHandler(uc *appcore.UseCase, p *presenter.AuthPresenter) *Handler {
	return &Handler{useCase: uc, presenter: p}
}

// NewHandlerFromInterface is used in tests to inject a stubbed use case.
func NewHandlerFromInterface(uc UseCaseInterface, p *presenter.AuthPresenter) *Handler {
	return &Handler{useCase: uc, presenter: p}
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}
	out, err := h.useCase.Register(c.Request.Context(), appcore.RegisterInput{
		Email: req.Email, Password: req.Password, DisplayName: req.DisplayName,
	})
	if errors.Is(err, domain.ErrEmailExists) {
		httpbase.Error(c, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "registration failed")
		return
	}
	h.presenter.Tokens(c, http.StatusCreated, out)
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}
	out, err := h.useCase.Login(c.Request.Context(), appcore.LoginInput{
		Email: req.Email, Password: req.Password,
	})
	if errors.Is(err, domain.ErrInvalidCredentials) {
		httpbase.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "login failed")
		return
	}
	h.presenter.Tokens(c, http.StatusOK, out)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}
	out, err := h.useCase.Refresh(c.Request.Context(), req.RefreshToken)
	if errors.Is(err, domain.ErrInvalidToken) {
		httpbase.Error(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "refresh failed")
		return
	}
	h.presenter.Tokens(c, http.StatusOK, out)
}

func (h *Handler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}
	if err := h.useCase.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "logout failed")
		return
	}
	httpbase.Success(c, http.StatusOK, gin.H{"message": "logged out"})
}
