package data

import (
	"database/sql"
	"errors"
	"greenlight.abdulalsh.com/internal/validator"
	"strings"
	"time"
)

var (
	AllowedLanguages = []string{"ar", "en"}
)

// note that all the fields are exported so they are visible to encoding/json package
// we added a struct tag to each field to be snake_case casing style
type Category struct {
	ID        int64     `json:"id"`    // Unique integer ID for the category
	CreatedAt time.Time `json:"-"`     // Timestamp for when the category is added to our database
	Title     string    `json:"title"` // category title
	Language  string    `json:"language"`
	Image     string    `json:"image"`
	Version   int32     `json:"-"` // The version number starts at 1 and will be incremented each its updated
}

func ValidateLanguage(v *validator.Validator, language string) {
	v.Check(validator.PermittedValues(language, AllowedLanguages...), "language", language+" not an allowed language")
}
func ValidateCategory(v *validator.Validator, category *Category) {
	v.Check(category.Title != "", "title", "must be provided")
	v.Check(len(category.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(category.Language != "", "language", "must be provided")
	v.Check(category.Image != "", "image", "must contain an image")
	v.Check(validator.PermittedValues(category.Language, AllowedLanguages...), "language", category.Language+" not an allowed language")
}

type CategoryModel struct {
	DB *sql.DB
}

const (
	insertCategoryQuery = `INSERT INTO categories(version) values (1) RETURNING id, created_at, version`

	insertCategoryTranslationQuery = `INSERT INTO category_translations (category_id,language_id,translation,image) values ($1,(SELECT id FROM languages WHERE code = $2),$3,$4)`

	// The UpsertTranslation method uses PostgreSQL’s INSERT ... ON CONFLICT syntax to perform an upsert.
	//This will insert a new translation if it does not exist or update the existing one if it does.
	//This method is more efficient and preferred if your database supports it.
	updateCategoryTranslationQuert = `
	INSERT INTO category_translations (category_id, language_id, translation,image)
	VALUES ($1, (SELECT id FROM languages WHERE code = $2), $3,$4)
	ON CONFLICT (category_id, language_id)
	DO UPDATE SET translation = EXCLUDED.translation, image = EXCLUDED.image;`

	getQuery = `
		SELECT
			c.id AS category_id,
			c.created_at,
			c.version,
			l.code AS language_code,
			ct.translation,
			ct.image

		FROM
			categories c
		JOIN
			category_translations ct ON c.id = ct.category_id
		JOIN
			languages l ON ct.language_id = l.id
		WHERE
			c.id = $1 AND l.code = $2;`

	getAllQuery = `
		SELECT
	c.id AS category_id,
		c.created_at,
		c.version,
		l.code AS language_code,
		ct.translation,
		ct.image
	FROM
	categories c
	JOIN
	category_translations ct ON c.id = ct.category_id
	JOIN
	languages l ON ct.language_id = l.id
	WHERE
	l.code = $1;`

	deleteQuery = `DELETE FROM categories WHERE id=$1`
)

func (m CategoryModel) Insert(category *Category) error {
	ctx, cancel := createContext()
	defer cancel()
	err := m.DB.QueryRowContext(ctx, insertCategoryQuery).Scan(&category.ID, &category.CreatedAt, &category.Version)
	if err != nil {
		return err
	}
	ctx, cancel = createContext()
	defer cancel()
	_, err = m.DB.ExecContext(ctx, insertCategoryTranslationQuery, category.ID, category.Language, category.Title, category.Image)
	return err
}

func (m CategoryModel) Update(category *Category) error {
	ctx, cancel := createContext()
	defer cancel()
	_, err := m.DB.ExecContext(ctx, updateCategoryTranslationQuert, category.ID, category.Language, category.Title, category.Image)
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

func (m CategoryModel) Get(id int64, language string) (*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	category := Category{}
	ctx, cancel := createContext()
	defer cancel()
	err := m.DB.QueryRowContext(ctx, getQuery, id, language).Scan(&category.ID, &category.CreatedAt, &category.Version, &category.Language, &category.Title, &category.Image)
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
	ctx, cancel := createContext()
	defer cancel()

	result, err := m.DB.ExecContext(ctx, deleteQuery, id)
	if err != nil {
		if strings.Contains(err.Error(), "Cannot delete the default category") {
			return ErrCantDeleteDefaultCategory
		}
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

	ctx, cancel := createContext()
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, getAllQuery, language)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var ct Category
		if err := rows.Scan(&ct.ID, &ct.CreatedAt, &ct.Version, &ct.Language, &ct.Title, &ct.Image); err != nil {
			return nil, err
		}
		categories = append(categories, ct)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
