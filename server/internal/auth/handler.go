package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sonpham/vielish/server/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	tokens, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName)
	if errors.Is(err, ErrEmailExists) {
		response.Error(c, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "registration failed")
		return
	}

	response.Success(c, http.StatusCreated, tokens)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	tokens, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		response.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "login failed")
		return
	}

	response.Success(c, http.StatusOK, tokens)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	tokens, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	response.Success(c, http.StatusOK, tokens)
}

func (h *Handler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, http.StatusInternalServerError, "logout failed")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "logged out"})
}
