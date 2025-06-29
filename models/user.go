package models

import (
	"database/sql"
	"errors"
)

type User struct {
	ID           int
	Email        string
	PasswordHash string
	Role         string
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	query := `SELECT id, email, password_hash, role FROM user_login WHERE email = ?`
	row := db.QueryRow(query, email)

	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}
