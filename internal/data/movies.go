package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"greenlight.abdulalsh.com/internal/validator"
	"time"
)

// note that all the fields are exported so they are visible to encoding/json package
// we added a struct tag to each field to be snake_case casing style
type Movie struct {
	ID        int64     `json:"id"`                // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`                 // Timestamp for when the movie is added to our database
	Title     string    `json:"title"`             // Movie title
	Year      int32     `json:"year,omitempty"`    // Movie release year
	Runtime   Runtime   `json:"runtime,omitempty"` // Movie runtime (in minutes)
	Genres    []string  `json:"genres,omitempty"`  // Slice of genres for the movie (romance, comedy, etc.)
	Version   int32     `json:"version"`           // The version number starts at 1 and will be incremented each
	// time the movie information is updated
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB *sql.DB
}

// Add a placeholder method for inserting a new record in the movies table.
func (m MovieModel) Insert(movie *Movie) error {
	// Define the SQL query for inserting a new record in the movies table and returning
	// the system-generated data.
	query := `
        INSERT INTO movies (title, year, runtime, genres) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	// Use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameter and scanning the system-
	// generated id, created_at and version values into the movie struct.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Add a placeholder method for fetching a specific record from the movies table.
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	stmt := `SELECT  id, created_at, title, year, runtime, genres, version FROM movies where id =$1`
	movie := Movie{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

// Add a placeholder method for updating a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	stmt := `UPDATE movies SET title =$1, year=$2, runtime =$3, genres=$4, version= version+1 WHERE id=$5 AND version=$6 RETURNING version`
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Add a placeholder method for deleting a specific record from the movies table.
func (m MovieModel) Delete(id int64) error {
	stmt := `DELETE FROM movies WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
func (m MovieModel) GetAll(title string, generes []string, filters Filters) ([]*Movie, error) {
	/*
		This SQL query is designed so that each of the filters behaves like it is ‘optional’. For example, the condition
		(LOWER(title) = LOWER($1) OR $1 = '') will evaluate as true if the placeholder parameter $1 is a case-insensitive
		match for the movie title or the placeholder parameter equals ''. So this filter condition will essentially be
		‘skipped’ when movie title being searched for is the empty string "".

		The (genres @> $2 OR $2 = '{}') condition works in the same way. The @> symbol is the ‘contains’ operator for
		PostgreSQL arrays, and this condition will return true if each value in the placeholder parameter $2 appears in the
		database genres field or the placeholder parameter contains an empty array.
	*/

	/*
		That looks pretty complicated at first glance, so let’s break it down and explain what’s going on.

		The to_tsvector('simple', title) function takes a movie title and splits it into lexemes. We specify the simple configuration, which means that the lexemes are just lowercase versions of the words in the title†. For example, the movie title "The Breakfast Club" would be split into the lexemes 'breakfast' 'club' 'the'.

		† Other ‘non-simple’ configurations may apply additional rules to the lexemes, such as the removal of common words or applying language-specific stemming.

		The plainto_tsquery('simple', $1) function takes a search value and turns it into a formatted query term that PostgreSQL full-text search can understand. It normalizes the search value (again using the simple configuration), strips any special characters, and inserts the and operator & between the words. As an example, the search value "The Club" would result in the query term 'the' & 'club'.

		The @@ operator is the matches operator. In our statement we are using it to check whether the generated query term matches the lexemes. To continue the example, the query term 'the' & 'club' will match rows which contain both lexemes 'the' and 'club'.
	*/

	/*
	   // Add an ORDER BY clause and interpolate the sort column and direction. Importantly
	   // notice that we also include a secondary sort on the movie ID to ensure a
	   // consistent ordering.
	*/
	stmt := fmt.Sprintf(`
        SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') 
		AND (genres @> $2 OR $2 = '{}')     
		ORDER BY %s %s,id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// As our SQL query now has quite a few placeholder parameters, let's collect the
	// values for the placeholders in a slice. Notice here how we call the limit() and
	// offset() methods on the Filters struct to get the appropriate values for the
	// LIMIT and OFFSET clauses.
	args := []any{title, pq.Array(generes), filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	//defer rows.close to ensure resultset is closed
	defer rows.Close()
	movies := []*Movie{}
	for rows.Next() {
		var movie Movie
		//scan the value into the struct
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}
	//when the rows next loop is done, call rows.err to retrive any errors
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
}
