package server

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func (s *Server) scanMovies(c *gin.Context) {
	dir := os.Getenv("MOVIES_DIR")

	if dir == "" {
		log.Println("No movies directory provided")
		c.Status(http.StatusNoContent)
		return
	}

	err := s.db.ScanDirectory(dir)
	if err != nil {
		log.Println("Failed while trying to scan local movies directory:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Scanning Failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Movies scanned successfully!"})
}

func (s *Server) getMovies(c *gin.Context) {
	movies, err := s.db.GetAllMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err})
		return
	}

	c.JSON(http.StatusOK, movies)
}
