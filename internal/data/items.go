package data

import (
	"database/sql"
	"errors"
	"greenlight.abdulalsh.com/internal/validator"
	"strings"
	"time"
)

var (
	ErrDuplicateItemTranslation = errors.New("duplicate item translation")
)

// Item represents an item along with its translation.
type Item struct {
	ID         int64     `json:"id"`          // Unique integer ID for the item
	CategoryID int64     `json:"category_id"` // Category ID for the item
	CreatedAt  time.Time `json:"-"`           // Timestamp for when the item is added to our database
	Name       string    `json:"name"`        // Item name
	Image      string    `json:"image"`       // Item image
	Language   string    `json:"language"`    // Language code
}

// ValidateItem validates the item fields.
func ValidateItem(v *validator.Validator, item *Item) {
	v.Check(item.Name != "", "name", "must be provided")
	v.Check(len(item.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(item.Language != "", "language", "must be provided")
	v.Check(item.Image != "", "image", "must contain an image")
	v.Check(validator.PermittedValues(item.Language, AllowedLanguages...), "language", item.Language+" not an allowed language")
}

type ItemModel struct {
	DB *sql.DB
}

const (
	insertItemQuery            = `INSERT INTO items(category_id) VALUES ($1) RETURNING id, created_at`
	insertItemTranslationQuery = `INSERT INTO item_translations (item_id, language_id, translation, image) VALUES ($1, (SELECT id FROM languages WHERE code = $2), $3, $4)`
	upsertItemTranslationQuery = `INSERT INTO item_translations (item_id, language_id, translation, image) VALUES ($1, (SELECT id FROM languages WHERE code = $2), $3, $4) ON CONFLICT (item_id, language_id) DO UPDATE SET translation = EXCLUDED.translation, image = EXCLUDED.image`
	updateCategoryId           = "update items set category_id = $1 where id = $2"
	getItemQuery               = `
		SELECT
			i.id,
			i.category_id,
			i.created_at,
			it.translation AS title,
			it.image,
			l.code AS language
		FROM
			items i
		JOIN
			item_translations it ON i.id = it.item_id
		JOIN
			languages l ON it.language_id = l.id
		WHERE
			i.id = $1 AND l.code = $2`
	getAllItemsQuery = `
		SELECT
			i.id,
			i.category_id,
			i.created_at,
			it.translation AS title,
			it.image,
			l.code AS language
		FROM
			items i
		JOIN
			item_translations it ON i.id = it.item_id
		JOIN
			languages l ON it.language_id = l.id
		WHERE
			l.code = $1`
	deleteItemQuery = `DELETE FROM items WHERE id = $1`
)

func (m ItemModel) Insert(item *Item) error {
	ctx, cancel := createContext()
	defer cancel()
	err := m.DB.QueryRowContext(ctx, insertItemQuery, item.CategoryID).Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return err
	}
	ctx, cancel = createContext()
	defer cancel()
	_, err = m.DB.ExecContext(ctx, insertItemTranslationQuery, item.ID, item.Language, item.Name, item.Image)
	return err
}

func (m ItemModel) Update(item *Item, updateCategory bool) error {
	ctx, cancel := createContext()
	defer cancel()
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(upsertItemTranslationQuery, item.ID, item.Language, item.Name, item.Image)
	if err != nil {

		if strings.Contains(err.Error(), `item_translations_item_id_language_id_key`) && strings.Contains(err.Error(), "duplicate") {
			return ErrDuplicateItemTranslation
		}

		return err
	}
	if updateCategory {

		_, err = tx.Exec(updateCategoryId, item.CategoryID, item.ID)
		if err != nil {
			if err.Error() == "pq: insert or update on table \"items\" violates foreign key constraint \"items_category_id_fkey\"" {
				return ErrCategoryDoesntExist
			}
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (m ItemModel) Get(id int64, language string) (*Item, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	item := Item{}
	ctx, cancel := createContext()
	defer cancel()
	err := m.DB.QueryRowContext(ctx, getItemQuery, id, language).Scan(&item.ID, &item.CategoryID, &item.CreatedAt, &item.Name, &item.Image, &item.Language)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (m ItemModel) GetAll(language string) ([]Item, error) {
	ctx, cancel := createContext()
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, getAllItemsQuery, language)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.CreatedAt, &item.Name, &item.Image, &item.Language); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (m ItemModel) Delete(id int64) error {
	ctx, cancel := createContext()
	defer cancel()

	result, err := m.DB.ExecContext(ctx, deleteItemQuery, id)
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
