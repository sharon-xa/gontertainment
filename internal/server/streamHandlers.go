package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) streamVideo(c *gin.Context) {
	movieID := c.Param("movie_id")
	if movieID == "" {
		log.Println("Movie path is required")
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Movie path is required"})
		return
	}

	moviePath, err := s.db.GetMoviePath(movieID)
	if err != nil {
		log.Printf("Movie with ID %s doesn't exists", movieID)
		c.JSON(http.StatusNotFound, gin.H{"msg": "Movie not found"})
		return
	}

	// Open the video file
	file, err := os.Open(moviePath)
	if err != nil {
		log.Println("Error opening video file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error opening video file"})
		return
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		log.Println("Error getting file information:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Error getting file information"})
		return
	}
	fileSize := fileInfo.Size()

	// Check if the client sent a Range request
	rangeHeader := c.Request.Header.Get("Range")
	fmt.Println("Range Header:", rangeHeader)
	if rangeHeader != "" {
		// Extract the byte range from the header
		rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
		byteRange := strings.Split(rangeHeader, "-")
		start, err := strconv.ParseInt(byteRange[0], 10, 64)
		if err != nil {
			log.Println("Invalid range start:", err)
			c.JSON(http.StatusBadRequest, "Invalid range")
			return
		}

		// Determine the end of the range
		var end int64
		if len(byteRange) > 1 && byteRange[1] != "" {
			end, err = strconv.ParseInt(byteRange[1], 10, 64)
			if err != nil {
				log.Println("Invalid range end:", err)
				c.JSON(http.StatusBadRequest, "Invalid range")
				return
			}
		} else {
			end = fileSize - 1
		}

		if end >= fileSize {
			end = fileSize - 1
		}

		// Set the appropriate headers for partial content
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
		c.Header("Content-Type", getFileContentType(moviePath))
		c.Status(http.StatusPartialContent)

		// Seek to the start of the requested range
		_, err = file.Seek(start, io.SeekStart)
		if err != nil {
			log.Println("Error seeking file:", err)
			return
		}

		// Stream the requested range using the helper function
		err = copyWithErrorHandling(c, file, end-start+1)
		if err != nil {
			return
		}

	} else {
		// No range request, serve the entire file
		c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
		c.Header("Content-Type", getFileContentType(moviePath))

		err = copyWithErrorHandling(c, file, fileSize)
		if err != nil {
			return
		}
	}
}

// Helper function to handle file streaming and error checking
func copyWithErrorHandling(c *gin.Context, file *os.File, bytesToCopy int64) error {
	_, err := io.CopyN(c.Writer, file, bytesToCopy)
	if err != nil {
		// Handle broken pipe errors gracefully
		if strings.Contains(err.Error(), "broken pipe") {
			log.Println("Client closed connection (broken pipe)")
			return err
		}
		log.Println("Error streaming video file:", err)
		return err
	}
	return nil
}

// Helper function to determine the Content-Type based on file extension
func getFileContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	default:
		return "application/octet-stream" // Default binary stream for unknown file types
	}
}
