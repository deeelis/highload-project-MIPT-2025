package middleware_controller

import (
	"api_gateway/internal/config"
	"api_gateway/internal/usecases/auth_usecase"
	"api_gateway/logger"
	"context"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
)

func AuthMiddleware(cfg *config.AuthConfig, log *slog.Logger) (gin.HandlerFunc, error) {
	const op = "middleware_controller.AuthMiddleware"
	log = log.With(slog.String("op", op))
	log.Info("initializing auth middleware")

	uc, err := auth_usecase.NewAuthUsecase(cfg, log)
	if err != nil {
		log.Error("failed to create auth usecase", logger.Err(err))
		return nil, err
	}

	return func(ctx *gin.Context) {
		startTime := time.Now()
		requestID := ctx.GetString("request_id")
		method := ctx.Request.Method
		path := ctx.Request.URL.Path

		log := log.With(
			slog.String("method", method),
			slog.String("path", path),
			slog.String("request_id", requestID),
		)

		log.Info("auth middleware processing request")

		token := ctx.GetHeader("Authorization")
		if token == "" {
			log.Warn("missing authorization header",
				slog.String("client_ip", ctx.ClientIP()))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			return
		}

		validateCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
		defer cancel()

		userID, err := uc.ValidateToken(validateCtx, token)
		if err != nil {
			log.Warn("token validation failed",
				logger.Err(err),
				slog.String("client_ip", ctx.ClientIP()),
				slog.Duration("duration", time.Since(startTime)))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		log.Info("token validation successful",
			slog.String("user_id", userID),
			slog.Duration("duration", time.Since(startTime)))

		ctx.Set("userID", userID)
		ctx.Set("logger", log.With(slog.String("user_id", userID)))
		ctx.Next()
	}, nil
}
