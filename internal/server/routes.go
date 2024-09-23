package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gontertainment/internal/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	r.Use(middleware.SetupCORS())

	r.POST("/scan", s.scanMovies)
	r.GET("/movies", s.getMovies)
	r.GET("/movie/:movie_id", s.streamVideo)

	r.GET("/health", s.healthHandler)

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
