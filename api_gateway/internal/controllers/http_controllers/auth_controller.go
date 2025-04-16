package http_controllers

import (
	"api_gateway/internal/config"
	"api_gateway/internal/domain/errors"
	"api_gateway/internal/usecases"
	"api_gateway/internal/usecases/auth_usecase"
	"api_gateway/logger"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	cfg    *config.AuthConfig
	authUC usecases.AuthUsecase
	log    *slog.Logger
}

func NewAuthController(cfg *config.AuthConfig, log *slog.Logger) (*AuthController, error) {
	uc, err := auth_usecase.NewAuthUsecase(cfg, log)
	if err != nil {
		return nil, err
	}
	return &AuthController{cfg: cfg, authUC: uc, log: log}, nil
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Name     string `json:"name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Warn("invalid request body", logger.Err(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userId, err := c.authUC.Register(ctx.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		c.log.Error("registration failed", logger.Err(err))
		ctx.JSON(errorToStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, userId)
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Warn("invalid request body", logger.Err(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tokens, err := c.authUC.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.log.Warn("login failed", logger.Err(err))
		ctx.JSON(errorToStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, tokens)
}

func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Warn("invalid request body", logger.Err(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tokens, err := c.authUC.RefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		c.log.Warn("refresh token failed", logger.Err(err))
		ctx.JSON(errorToStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func errorToStatusCode(err error) int {
	switch err {
	case errors.ErrInvalidCredentials:
		return http.StatusUnauthorized
	case errors.ErrUserAlreadyExists:
		return http.StatusConflict
	case errors.ErrUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
