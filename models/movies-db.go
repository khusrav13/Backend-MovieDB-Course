package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type DBModel struct {
	DB *sql.DB
}

// Get / Return only one movie and error, if any
func (p *DBModel) Get(id int) (*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `select id, title, description, year, release_date, rating, runtime, mpaa_rating,
				created_at, updated_at from movies where id = $1
	`

	row := p.DB.QueryRowContext(ctx, query, id)
	var movie Movie
	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.Year,
		&movie.ReleaseDate,
		&movie.Rating,
		&movie.Runtime,
		&movie.MPAARating,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// get the genres

	query = `SELECT mg.id, mg.movie_id, mg.genre_id, g.genre_name
	FROM movies_genres mg
	left join genres g on (g.id = mg.genre_id)
	WHERE mg.movie_id = $1`

	rows, _ := p.DB.QueryContext(ctx, query, id)
	defer rows.Close()

	genres := make(map[int]string)
	for rows.Next() {
		var mg MovieGenre
		err := rows.Scan(
			&mg.ID,
			&mg.MovieId,
			&mg.GenreId,
			&mg.Genre.GenreName,
		)
		if err != nil {
			return nil, err
		}
		genres[mg.ID] = mg.Genre.GenreName
	}

	movie.MovieGenre = genres

	return &movie, nil
}

// All / Return all movies and error, if any
func (p *DBModel) All(genre ...int) ([]*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	where := ""
	if len(genre) > 0 {
		where = fmt.Sprintf("WHERE id in (SELECT movie_id FROM movies_genres WHERE genre_id = %d)", genre[0])
	}

	query := fmt.Sprintf(`select id, title, description, year, release_date, runtime, rating,  mpaa_rating,
				created_at, updated_at from movies %s order by id`, where)

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*Movie

	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Description,
			&movie.Year,
			&movie.ReleaseDate,
			&movie.Runtime,
			&movie.Rating,
			&movie.MPAARating,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// get the genres

		qenreQuery := `SELECT mg.id, mg.movie_id, mg.genre_id, g.genre_name
	FROM movies_genres mg
	left join genres g on (g.id = mg.genre_id)
	WHERE mg.movie_id = $1`

		genreRows, _ := p.DB.QueryContext(ctx, qenreQuery, movie.ID)

		genres := make(map[int]string)
		for genreRows.Next() {
			var mg MovieGenre
			err := genreRows.Scan(
				&mg.ID,
				&mg.MovieId,
				&mg.GenreId,
				&mg.Genre.GenreName,
			)
			if err != nil {
				return nil, err
			}
			genres[mg.ID] = mg.Genre.GenreName
		}
		genreRows.Close()

		movie.MovieGenre = genres
		movies = append(movies, &movie)
	}

	return movies, nil
}

func (p *DBModel) GenresAll() ([]*Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, genre_name, created_at, updated_at FROM genres ORDER BY id`

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*Genre

	for rows.Next() {
		var g Genre
		err := rows.Scan(
			&g.ID,
			&g.GenreName,
			&g.CreatedAt,
			&g.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		genres = append(genres, &g)
	}
	return genres, nil
}

func (m *DBModel) InsertMovie(movie Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	var count int
	query := `SELECT MAX(id) FROM movies;`
	if err := m.DB.QueryRow(query).Scan(&count); err != nil {
		log.Println(err)
		return err
	}

	newDB := `INSERT INTO movies(id, title, description, year, release_date, runtime, rating, mpaa_rating, created_at, updated_at)
		VALUES ($1, $2, $3, $4,$5, $6, $7, $8, $9, $10);`

	_, err := m.DB.ExecContext(ctx, newDB,

		movie.ID,
		movie.Title,
		movie.Description,
		movie.Year,
		movie.ReleaseDate,
		movie.Runtime,
		movie.Rating,
		movie.MPAARating,
		movie.CreatedAt,
		movie.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}
