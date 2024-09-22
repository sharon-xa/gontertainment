package database

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Movie struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
	FileSize int64  `json:"fileSize"`
	Format   string `json:"format"`

	Overview    string `json:"overview"`
	PosterURL   string `json:"posterURL"`
	ReleaseDate string `json:"release_date"`
}

type TMDbMovieResponse struct {
	Results []struct {
		Title      string `json:"title"`
		Overview   string `json:"overview"`
		PosterPath string `json:"poster_path"`
	} `json:"results"`
}

// Function to fetch movie details from TMDb by title
func getMovieDetailsFromTMDb(title string) (TMDbMovieResponse, error) {
	urlTitle := strings.ReplaceAll(title, " ", "%20")
	apiKey := os.Getenv("TMDB_API_KEY")
	baseURL := "https://api.themoviedb.org/3/search/movie"
	tmdbURL := fmt.Sprintf("%s?api_key=%s&query=%s", baseURL, apiKey, urlTitle)

	resp, err := http.Get(tmdbURL)
	if err != nil {
		return TMDbMovieResponse{}, err
	}
	defer resp.Body.Close()

	var movieResponse TMDbMovieResponse
	err = json.NewDecoder(resp.Body).Decode(&movieResponse)
	if err != nil {
		log.Println("ERROR:", err)
		return TMDbMovieResponse{}, err
	}

	return movieResponse, nil
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

		format := strings.ToLower(filepath.Ext(info.Name()))

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

		tmdbResponse, tmdbErr := getMovieDetailsFromTMDb(title)
		if tmdbErr != nil {
			log.Printf("Error fetching TMDb data for %s: %v", title, tmdbErr)
		} else if len(tmdbResponse.Results) > 0 {
			tmdbMovie := tmdbResponse.Results[0]
			movie.Overview = tmdbMovie.Overview
			movie.PosterURL = "https://image.tmdb.org/t/p/w500" + tmdbMovie.PosterPath
		}

		_, dbErr := tx.Exec(
			`INSERT INTO movies (title, file_name, file_path, file_size, format, overview, poster_url)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (file_path) DO UPDATE
			SET
				title = EXCLUDED.title,
			  file_name = EXCLUDED.file_name,
				file_path = EXCLUDED.file_path,
				file_size = EXCLUDED.file_size,
    		format = EXCLUDED.format,
    		overview = EXCLUDED.overview,
    		poster_url = EXCLUDED.poster_url;`,
			movie.Title,
			movie.FileName,
			movie.FilePath,
			movie.FileSize,
			movie.Format,
			movie.Overview,
			movie.PosterURL,
			time.Now(),
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
	rows, err := s.db.Query(
		"SELECT id, title, file_name, file_path, file_size, format, overview, poster_url FROM movies",
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to Start a Query Statement: %w", err)
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.FileName,
			&movie.FilePath,
			&movie.FileSize,
			&movie.Format,
			&movie.Overview,
			&movie.PosterURL,
		)
		if err != nil {
			log.Println(err)
			continue
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (s *service) GetMoviePath(movieID string) (string, error) {
	row := s.db.QueryRow(
		"SELECT file_path FROM movies WHERE id = $1",
		movieID,
	)

	var moviePath string

	err := row.Scan(&moviePath)
	if err != nil {
		log.Println("Failed while trying to retrive movie path from db\n", err)
		return "", fmt.Errorf("couldn't get movie path")
	}

	return moviePath, nil
}

// func (s *service) SetMovieReleaseYear(movieReleaseYear string) error {
//
// }
