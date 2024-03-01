package database

import (
	"time"
)

type Revocation struct {
	Token     string    `json:"token"`
	RevokedAt time.Time `json:"revoked_at"`
}

// store new revoked token into db
func (db *DB) RevokeToken(token string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	revocation := Revocation{
		Token:     token,
		RevokedAt: time.Now().UTC(),
	}
	dbStructure.Revocations[token] = revocation

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}

// check if token exists in revoked map in db
func (db *DB) IsTokenRevoked(token string) (bool, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}
	// check if current refresh token exsits
	revocation, ok := dbStructure.Revocations[token]
	if !ok {
		return false, nil
	}
	// invalid entry if timing is zero
	if revocation.RevokedAt.IsZero() {
		return false, nil
	}

	return true, nil
}
