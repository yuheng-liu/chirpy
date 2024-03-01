package database

import (
	"errors"
)

// struct used for storing user data
type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	// should store in hashed value
	HashedPassword string `json:"hashed_password"`
}

var ErrAlreadyExists = errors.New("already exists")

func (db *DB) CreateUser(email, hashedPassword string) (User, error) {
	// check if user already exists in db
	if _, err := db.GetUserByEmail(email); !errors.Is(err, ErrNotExist) {
		return User{}, ErrAlreadyExists
	}
	// load db and check error
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	// simple id incremental value
	id := len(dbStructure.Users) + 1
	user := User{
		ID:             id,
		Email:          email,
		HashedPassword: hashedPassword,
	}
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) GetUser(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	// check if user exists
	user, ok := dbStructure.Users[id]
	if !ok {
		// cant find user, return error
		return User{}, ErrNotExist
	}

	return user, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	// check if user email exists
	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}
	// can't find user, return error
	return User{}, ErrNotExist
}

func (db *DB) UpdateUser(id int, email, hashedPassword string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	// check if user exists
	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}
	// replace old user entry with new values
	user.Email = email
	user.HashedPassword = hashedPassword
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
