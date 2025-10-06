// internal/handler/auth.go
package handler

import (
	"auth-service/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c echo.Context) error {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	req := new(Request)
	if err := c.Bind(req); err != nil {
		return echo.ErrBadRequest
	}

	if err := h.authService.Register(c.Request().Context(), req.Email, req.Password); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "User registered"})
}

func (h *AuthHandler) Login(c echo.Context) error {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	req := new(Request)
	if err := c.Bind(req); err != nil {
		return echo.ErrBadRequest
	}

	token, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return echo.ErrUnauthorized
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}
