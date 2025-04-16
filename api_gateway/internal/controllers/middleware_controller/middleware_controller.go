package middleware_controller

import (
	"api_gateway/internal/config"
	"api_gateway/internal/usecases/auth_usecase"
	"api_gateway/logger"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

func AuthMiddleware(cfg *config.AuthConfig, log *slog.Logger) (gin.HandlerFunc, error) {
	uc, err := auth_usecase.NewAuthUsecase(cfg, log)
	if err != nil {
		return nil, err
	}
	return func(ctx *gin.Context) {

		token := ctx.GetHeader("Authorization")
		if token == "" {
			log.Warn("missing authorization header")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, err := uc.ValidateToken(ctx.Request.Context(), token)
		if err != nil {
			log.Warn("invalid token", logger.Err(err))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		ctx.Set("userID", userID)
		ctx.Next()
	}, nil
}
