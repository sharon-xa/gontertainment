package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Movie struct {
	ID       int
	Title    string
	FileName string
	FilePath string
	FileSize int64
	Format   string
}

func (s *service) ScanDirectory(dir string) error {
	videoExtensions := map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".avi":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get the file extension
		format := strings.ToLower(filepath.Ext(info.Name()))

		// Check if the file is a video file based on its extension
		if !videoExtensions[format] {
			log.Printf("Skipping non-video file: %s", info.Name())
			return nil
		}

		title := strings.Split(info.Name(), ".")[0]

		movie := Movie{
			Title:    title,
			FileName: info.Name(),
			FilePath: path,
			FileSize: info.Size(),
			Format:   format,
		}

		_, dbErr := tx.Exec(
			`INSERT INTO movies (title, file_name, file_path, file_size, format) 
      VALUES ($1, $2, $3, $4, $5) 
      ON CONFLICT (file_path) DO NOTHING`,
			movie.Title,
			movie.FileName,
			movie.FilePath,
			movie.FileSize,
			movie.Format,
		)
		if dbErr != nil {
			log.Printf("Error inserting movie %s: %v", movie.Title, dbErr)
			tx.Rollback()
			return dbErr
		}

		log.Printf("Added %s to database", movie.Title)
		return nil
	})
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	return nil
}

func (s *service) GetAllMovies() ([]Movie, error) {
	rows, err := s.db.Query("SELECT id, title, file_name, format, file_size FROM movies")
	if err != nil {
		return nil, fmt.Errorf("Failed to Start a Query Statement: %w", err)
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		err := rows.Scan(&movie.ID, &movie.Title, &movie.FileName, &movie.Format, &movie.FileSize)
		if err != nil {
			log.Println(err)
			continue
		}
		movies = append(movies, movie)
	}

	return movies, nil
}
