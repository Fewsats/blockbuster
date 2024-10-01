package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"log/slog"

	"github.com/gin-gonic/gin"

	"html/template"

	"github.com/fewsats/blockbuster/auth"
	"github.com/fewsats/blockbuster/config"
	"github.com/fewsats/blockbuster/video"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

//go:embed frontend
var frontendFS embed.FS

type Server struct {
	router    *gin.Engine
	logger    *slog.Logger
	cfg       *config.Config
	auth      *auth.Controller
	video     *video.Controller
	templates *template.Template
}

func NewServer(logger *slog.Logger, cfg *config.Config, authCtrl *auth.Controller, videoCtrl *video.Controller) (*Server, error) {
	router := gin.New()
	router.Use(gin.Recovery())

	// Set up CORS middleware
	// TODO(pol) is it ok to allow all origins?
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.ExposeHeaders = []string{"Content-Length", "Content-Disposition", "Content-Type", "X-Requested-With", "Set-Cookie", "Www-authenticate"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Set up session middleware
	store := cookie.NewStore([]byte(cfg.Auth.SessionSecret))
	router.Use(sessions.Sessions("mysession", store))

	// Parse HTML templates
	tmpl, err := template.ParseFS(frontendFS, "frontend/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML templates: %w", err)
	}

	s := &Server{
		router:    router,
		logger:    logger,
		cfg:       cfg,
		auth:      authCtrl,
		video:     videoCtrl,
		templates: tmpl,
	}

	s.setupRoutes()

	return s, nil
}

func (s *Server) setupRoutes() {
	// Serve static files
	staticFS, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		s.logger.Error("Failed to create sub-filesystem for static files", "error", err)
		return
	}
	s.router.StaticFS("/static", http.FS(staticFS))

	// Use embedded templates
	s.router.SetHTMLTemplate(s.templates)

	s.router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	s.router.GET("/faq", func(c *gin.Context) {
		c.HTML(http.StatusOK, "faq.html", nil)
	})
	s.router.GET("/profile", func(c *gin.Context) {
		c.HTML(http.StatusOK, "profile.html", nil)
	})

	s.auth.RegisterPublicRoutes(s.router)
	s.video.RegisterPublicRoutes(s.router)

	s.video.RegisterL402Routes(s.router)

	// Routes protected by auth middleware
	s.auth.RegisterAuthMiddleware(s.router)
	s.auth.RegisterProtectedRoutes(s.router)
	s.video.RegisterProtectedRoutes(s.router)

}

func (s *Server) Run() error {
	return s.router.Run(fmt.Sprintf(":%d", s.cfg.Port))
}
