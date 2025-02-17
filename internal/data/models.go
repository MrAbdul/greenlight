package data

import (
	"database/sql"
)

//As an additional step, we’re going to wrap our MovieModel in a parent Models struct. Doing this is totally optional,
//but it has the benefit of giving you a convenient single ‘container’ which can hold and represent all your database models as your application grows.

// Create a Models struct which wraps the MovieModel. We'll add other models to this,
// like a UserModel and PermissionModel, as our build progresses.
type Models struct {
	Movies        MovieModel
	Users         UserModel
	Token         TokenModel
	Permissions   PermissionModel
	CategoryModel CategoryModel
	ItemModel     ItemModel
}

// For ease of use, we also add a New() method which returns a Models struct containing
// the initialized MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:        MovieModel{DB: db},
		Users:         UserModel{DB: db},
		Token:         TokenModel{DB: db},
		Permissions:   PermissionModel{DB: db},
		CategoryModel: CategoryModel{DB: db},
		ItemModel:     ItemModel{DB: db},
	}
}
