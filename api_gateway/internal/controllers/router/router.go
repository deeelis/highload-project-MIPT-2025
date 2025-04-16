package http_controllers

import (
	"api_gateway/internal/config"
	"api_gateway/internal/controllers/http_controllers"
	"api_gateway/internal/controllers/middleware_controller"
	"github.com/gin-gonic/gin"
	"log/slog"
)

func NewRouter(cfg *config.Config, log *slog.Logger) (*gin.Engine, error) {
	authController, err := http_controllers.NewAuthController(cfg.Auth, log)
	if err != nil {
		return nil, err
	}
	contentController, err := http_controllers.NewContentController(cfg, log)
	if err != nil {
		return nil, err
	}

	gin.SetMode(gin.DebugMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		log.Debug("incoming request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path))
		c.Next()
	})
	middlewareController, err := middleware_controller.AuthMiddleware(cfg.Auth, log)

	SetupRoutes(router, authController, contentController, middlewareController)

	return router, nil

}
func SetupRoutes(router *gin.Engine, authController *http_controllers.AuthController, contentController *http_controllers.ContentController, middlewareController gin.HandlerFunc) {
	router.Use(gin.Recovery())

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
		protected.GET("/content/:id/status", contentController.GetStatus)
	}
}
