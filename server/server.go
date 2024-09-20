package server

import (
	"fmt"
	"net/http"

	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/fewsats/blockbuster/auth"
	"github.com/fewsats/blockbuster/config"
	"github.com/fewsats/blockbuster/video"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

type Server struct {
	router *gin.Engine
	logger *slog.Logger
	cfg    *config.Config
	auth   *auth.Controller
	video  *video.Controller
}

func NewServer(logger *slog.Logger, cfg *config.Config, authCtrl *auth.Controller, videoCtrl *video.Controller) (*Server, error) {
	router := gin.New()
	router.Use(gin.Recovery())

	// Set up session middleware
	store := cookie.NewStore([]byte(cfg.Auth.SessionSecret))
	router.Use(sessions.Sessions("mysession", store))

	s := &Server{
		router: router,
		logger: logger,
		cfg:    cfg,
		auth:   authCtrl,
		video:  videoCtrl,
	}

	s.setupRoutes()

	return s, nil
}

func (s *Server) setupRoutes() {
	s.router.Static("/static", "./frontend")
	s.router.LoadHTMLGlob("frontend/*.html")

	s.router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
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
