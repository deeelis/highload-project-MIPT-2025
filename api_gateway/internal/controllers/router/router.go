package http_controllers

import (
	"api_gateway/internal/config"
	"api_gateway/internal/controllers/http_controllers"
	"api_gateway/internal/controllers/middleware_controller"
	"api_gateway/logger"
	"github.com/gin-gonic/gin"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

func NewRouter(cfg *config.Config, log *slog.Logger) (*gin.Engine, error) {
	const op = "http_controllers.NewRouter"
	log = log.With(slog.String("op", op))
	log.Info("initializing router")

	startTime := time.Now()
	authController, err := http_controllers.NewAuthController(cfg.Auth, log)
	if err != nil {
		log.Error("failed to create auth controller", logger.Err(err))
		return nil, err
	}
	contentController, err := http_controllers.NewContentController(cfg, log)
	if err != nil {
		log.Error("failed to create content controller", logger.Err(err))
		return nil, err
	}

	log.Debug("content controller initialized")

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.Info("running in production mode")
	} else {
		gin.SetMode(gin.DebugMode)
		log.Info("running in debug mode")
	}

	router := gin.New()
	router.Use(
		gin.Recovery(),
		requestLoggerMiddleware(log),
		requestIDMiddleware(),
	)

	middlewareController, err := middleware_controller.AuthMiddleware(cfg.Auth, log)
	if err != nil {
		log.Error("failed to create auth middleware", logger.Err(err))
		return nil, err
	}
	log.Debug("auth middleware initialized")

	SetupRoutes(router, authController, contentController, middlewareController)
	log.Info("router initialized successfully",
		slog.Duration("duration", time.Since(startTime)))

	return router, nil

}
func SetupRoutes(router *gin.Engine, authController *http_controllers.AuthController, contentController *http_controllers.ContentController, middlewareController gin.HandlerFunc) {
	public := router.Group("/auth")
	{
		public.POST("/register", authController.Register)
		public.POST("/login", authController.Login)
		public.POST("/refresh", authController.RefreshToken)
	}

	protected := router.Group("/")
	protected.Use(middlewareController)
	{
		protected.POST("/content/text", contentController.UploadText)
		protected.POST("/content/image", contentController.UploadImage)
		protected.GET("/content/:id", contentController.GetContent)
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "endpoint not found",
		})
	})
}

func requestLoggerMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestID := c.GetString("request_id")

		log.Info("incoming request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("ip", c.ClientIP()),
			slog.String("request_id", requestID))

		c.Next()

		status := c.Writer.Status()
		logEntry := log.With(
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(startTime)),
			slog.String("request_id", requestID),
		)

		if status >= http.StatusInternalServerError {
			logEntry.Error("server error")
		} else if status >= http.StatusBadRequest {
			logEntry.Warn("client error")
		} else {
			logEntry.Info("request completed")
		}
	}
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randString(8)
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
