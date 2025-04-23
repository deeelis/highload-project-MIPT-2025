package http_controllers

import (
	"api_gateway/internal/config"
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
	const op = "http_controllers.NewAuthController"
	log = log.With(slog.String("op", op))
	log.Info("initializing auth controller")

	uc, err := auth_usecase.NewAuthUsecase(cfg, log)
	if err != nil {
		log.Error("failed to create auth usecase", logger.Err(err))
		return nil, err
	}

	log.Info("auth controller initialized successfully")
	return &AuthController{cfg: cfg, authUC: uc, log: log}, nil
}

func (c *AuthController) Register(ctx *gin.Context) {
	const op = "http_controllers.AuthController.Register"
	log := c.log.With(
		slog.String("op", op),
		slog.String("method", ctx.Request.Method),
		slog.String("path", ctx.FullPath()),
	)

	log.Info("handling registration request")

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Name     string `json:"name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid registration request",
			logger.Err(err),
			slog.String("email", req.Email),
			slog.Int("name_length", len(req.Name)),
			slog.Bool("has_password", req.Password != ""))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	log.Debug("processing registration",
		slog.String("email", req.Email),
		slog.Int("name_length", len(req.Name)))

	tokenDetails, err := c.authUC.Register(ctx.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		log.Error("registration failed",
			logger.Err(err),
			slog.String("email", req.Email))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userId := tokenDetails.UserID

	log.Info("registration successful",
		slog.String("user_id", userId))
	ctx.JSON(http.StatusCreated, userId)
}

func (c *AuthController) Login(ctx *gin.Context) {
	const op = "http_controllers.AuthController.Login"
	log := c.log.With(
		slog.String("op", op),
		slog.String("method", ctx.Request.Method),
		slog.String("path", ctx.FullPath()),
	)

	log.Info("handling login request")

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid login request",
			logger.Err(err),
			slog.String("email", req.Email),
			slog.Bool("has_password", req.Password != ""))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	log.Debug("processing login",
		slog.String("email", req.Email))

	tokens, err := c.authUC.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		log.Warn("login failed",
			logger.Err(err),
			slog.String("email", req.Email))
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	log.Info("login successful",
		slog.String("email", req.Email))
	ctx.JSON(http.StatusOK, tokens)
}

func (c *AuthController) RefreshToken(ctx *gin.Context) {
	const op = "http_controllers.AuthController.RefreshToken"
	log := c.log.With(
		slog.String("op", op),
		slog.String("method", ctx.Request.Method),
		slog.String("path", ctx.FullPath()),
	)

	log.Info("handling refresh token request")

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid refresh token request",
			logger.Err(err),
			slog.Bool("has_token", req.RefreshToken != ""))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	log.Debug("processing refresh token",
		slog.Int("token_length", len(req.RefreshToken)))

	tokens, err := c.authUC.RefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		log.Warn("refresh token failed",
			logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info("refresh token successful")
	ctx.JSON(http.StatusOK, tokens)
}
