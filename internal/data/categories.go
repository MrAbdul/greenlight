package data

import (
	"context"
	"database/sql"
	"errors"
	"greenlight.abdulalsh.com/internal/validator"
	"strings"
	"time"
)

var (
	AllowedLanguages                = []string{"ar", "en"}
	ErrDublicateCategoryTranslation = errors.New("duplicate category translation")
)

// note that all the fields are exported so they are visible to encoding/json package
// we added a struct tag to each field to be snake_case casing style
type Category struct {
	ID        int64     `json:"id"`    // Unique integer ID for the category
	CreatedAt time.Time `json:"-"`     // Timestamp for when the category is added to our database
	Title     string    `json:"title"` // category title
	Language  string    `json:"language"`
	Version   int32     `json:"-"` // The version number starts at 1 and will be incremented each its updated
}

func ValidateLanguage(v *validator.Validator, language string) {
	v.Check(validator.PermittedValues(language, AllowedLanguages...), "language", language+" not an allowed language")
}
func ValidateCategory(v *validator.Validator, category *Category) {
	v.Check(category.Title != "", "title", "must be provided")
	v.Check(len(category.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(category.Language != "", "language", "must be provided")
	v.Check(validator.PermittedValues(category.Language, AllowedLanguages...), "language", category.Language+" not an allowed language")
}

type CategoryModel struct {
	DB *sql.DB
}

func (m CategoryModel) Insert(category *Category) error {
	query := `INSERT INTO categories(version) values (1) RETURNING id, created_at, version`
	querytran := `INSERT INTO category_translations (category_id,language_id,translation) values ($1,(SELECT id FROM languages WHERE code = $2),$3) `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query).Scan(&category.ID, &category.CreatedAt, &category.Version)
	if err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = m.DB.ExecContext(ctx, querytran, category.ID, category.Language, category.Title)
	return err
}
func (m CategoryModel) AddCategoryLanguage(category *Category) error {
	query := `SELECT created_at,version FROM categories WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, category.ID).Scan(&category.CreatedAt, &category.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	// The UpsertTranslation method uses PostgreSQL’s INSERT ... ON CONFLICT syntax to perform an upsert.
	//This will insert a new translation if it does not exist or update the existing one if it does.
	//This method is more efficient and preferred if your database supports it.
	upsert := `
	INSERT INTO category_translations (category_id, language_id, translation)
	VALUES ($1, (SELECT id FROM languages WHERE code = $2), $3)
	ON CONFLICT (category_id, language_id)
	DO UPDATE SET translation = EXCLUDED.translation;
`
	//stmt := `INSERT INTO category_translations(category_id, language_id, translation) values ($1,(SELECT id FROM languages WHERE code = $2),$3)`

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = m.DB.ExecContext(ctx, upsert, category.ID, category.Language, category.Title)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), `category_translations_category_id_language_id_key`) && strings.Contains(err.Error(), "duplicate"):
			return ErrDublicateCategoryTranslation
		default:
			return err
		}
	}
	return nil
}

// Add a placeholder method for fetching a specific record from the movies table.
func (m CategoryModel) Get(id int64, language string) (*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT
			c.id AS category_id,
			c.created_at,
			c.version,
			l.code AS language_code,
			ct.translation
		FROM
			categories c
		JOIN
			category_translations ct ON c.id = ct.category_id
		JOIN
			languages l ON ct.language_id = l.id
		WHERE
			c.id = $1 AND l.code = $2;
	`
	category := Category{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id, language).Scan(&category.ID, &category.CreatedAt, &category.Version, &category.Language, &category.Title)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &category, nil
}

// Add a placeholder method for deleting a specific record from the movies table.
func (m CategoryModel) Delete(id int64) error {
	stmt := `DELETE FROM categories WHERE id=$1`
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

func (m CategoryModel) GetAll(language string) ([]Category, error) {

	query := `
		SELECT
	c.id AS category_id,
		c.created_at,
		c.version,
		l.code AS language_code,
		ct.translation
	FROM
	categories c
	JOIN
	category_translations ct ON c.id = ct.category_id
	JOIN
	languages l ON ct.language_id = l.id
	WHERE
	l.code = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, language)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var ct Category
		if err := rows.Scan(&ct.ID, &ct.CreatedAt, &ct.Version, &ct.Language, &ct.Title); err != nil {
			return nil, err
		}
		categories = append(categories, ct)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
