package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordNotSet    = errors.New("password: no plaintext password set; call Set before persisting")
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Password  password `json:"-"`
	Email     string   `json:"email"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

// queryRower is satisfied by both *sql.DB and *sql.Tx, letting create run
// standalone or as part of an existing transaction.
type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// --- Create ---

func (u *UserStore) Create(ctx context.Context, user *User) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	return u.create(ctx, u.db, user)
}

func (u *UserStore) create(ctx context.Context, q queryRower, user *User) error {
	if user.Password.text == nil {
		return ErrPasswordNotSet
	}

	query := `
	INSERT INTO users (username, password, email)
	VALUES ($1, $2, $3) RETURNING id, created_at
	`

	err := q.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password.hash,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "users_email_key":
				return ErrDuplicateEmail
			case "users_username_key":
				return ErrDuplicateUsername
			}
		}
		return err
	}

	return nil
}

func (u *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, expiry time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		if err := u.create(ctx, tx, user); err != nil {
			return err
		}

		// create the user invite
		if err := u.createUserInvitation(ctx, tx, token, expiry, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (u *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, expiry time.Duration, userID int64) error {
	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	_, err := tx.ExecContext(ctx, query, hashToken(token), userID, time.Now().Add(expiry))
	if err != nil {
		return err
	}

	return nil
}

// --- Read ---

func (u *UserStore) GetAll(ctx context.Context) ([]User, error) {
	query := `
	SELECT id, username, email, created_at
	FROM users
	ORDER BY id
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := u.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (u *UserStore) GetByID(ctx context.Context, userID int64) (*User, error) {
	query := `
	SELECT id, username, email, created_at
	FROM users
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := u.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}

	}

	return user, nil
}

func (u *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
	SELECT u.id, u.username, u.email, u.created_at, u.is_active
	FROM users u
	JOIN user_invitations ui ON u.id = ui.user_id
	WHERE ui.token = $1 AND ui.expiry > $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken(token), time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

// --- Update ---

func (u *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		// find the user that this token belongs to
		user, err := u.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		// update the user
		user.IsActive = true
		if err := u.update(ctx, tx, user); err != nil {
			return err
		}

		// clean the invitation
		if err := u.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (u *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

// --- Delete ---

func (u *UserStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users where id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := u.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (u *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

// --- Helpers ---

// hashToken hashes an invitation token before storage. Unlike passwords,
// tokens are already high-entropy random values, so a fast, deterministic
// hash (sha256) is sufficient — there's no dictionary/brute-force risk to
// defend against with a slow hash like bcrypt, and a deterministic hash lets
// us look the row up directly by the hash of a presented token.
func hashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}
