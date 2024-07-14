package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"fmt"
	"greenlight.abdulalsh.com/internal/validator"
	"time"
	"unicode/utf8"
)

// Define constants for the token scope. For now we just define the scope "activation"
// but we'll add additional scopes later in the book.
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)
const (
	ActivationTokenLen     = 6
	AuthenticationTokenLen = 32
)

// Define a Token struct to hold the data for an individual token. This includes the
// plaintext and hashed versions of the token, associated user ID, expiry time and
// scope.
type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userID int64, ttl time.Duration, length int, scope string) (*Token, error) {
	// Create a Token instance containing the user ID, expiry, and scope information.
	// Notice that we add the provided ttl (time-to-live) duration parameter to the
	// current time to get the expiry time?
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Initialize a zero-valued byte slice with a length of 16 bytes.
	randomBytes := make([]byte, length)

	// Use the Read() function from the crypto/rand package to fill the byte slice with
	// random bytes from your operating system's CSPRNG. This will return an error if
	// the CSPRNG fails to function correctly.
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32-encoded string and assign it to the token
	// Plaintext field. This will be the token string that we send to the user in their
	// welcome email. They will look similar to this:
	//
	// Y3QMGX3PJ3WLRL2YRTQGQ6KRHU
	//
	// Note that by default base-32 strings may be padded at the end with the =
	// character. We don't need this padding character for the purpose of our tokens, so
	// we use the WithPadding(base32.NoPadding) method in the line below to omit them.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	if len(token.Plaintext) < length {
		return nil, fmt.Errorf("generated token is shorter than the required length")
	}
	token.Plaintext = token.Plaintext[:length]
	// Generate a SHA-256 hash of the plaintext token string. This will be the value
	// that we store in the `hash` field of our database table. Note that the
	// sha256.Sum256() function returns an *array* of length 32, so to make it easier to
	// work with we convert it to a slice using the [:] operator before storing it.
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

// Check that the plaintext token has been provided and is exactly 26 bytes long.
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string, scope string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	if scope == ScopeActivation {
		v.Check(utf8.RuneCountInString(tokenPlaintext) == ActivationTokenLen, "token", "must be 6 characters long")
	} else if scope == ScopeAuthentication {
		v.Check(utf8.RuneCountInString(tokenPlaintext) == AuthenticationTokenLen, "token", "must be 32 characters long")
	} else {
		v.AddError("token", "scope not defined")
	}
}

// Define the TokenModel type.
type TokenModel struct {
	DB *sql.DB
}

// The New() method is a shortcut which creates a new Token struct and then inserts the
// data in the tokens table.
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	var token *Token
	var err error
	if scope == ScopeActivation {
		token, err = generateToken(userID, ttl, 6, scope)
	} else if scope == ScopeAuthentication {
		token, err = generateToken(userID, ttl, 32, scope)
	} else {
		return nil, fmt.Errorf("scope must be defined")
	}
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

// Insert() adds the data for a specific token to the tokens table.
func (m TokenModel) Insert(token *Token) error {
	query := `
        INSERT INTO tokens (hash, user_id, expiry, scope) 
        VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// DeleteAllForUser() deletes all tokens for a specific user and scope.
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
        DELETE FROM tokens 
        WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

func (m TokenModel) GetAllForUser(userId int64) ([]*Token, error) {

	stmt := "SELECT hash,expiry,scope,user_id FROM tokens WHERE user_id = $1"

	var tokens []*Token
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		token := &Token{}
		err := rows.Scan(&token.Hash, &token.Expiry, &token.Scope)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tokens, nil

}
