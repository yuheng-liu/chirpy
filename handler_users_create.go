package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yuheng-liu/chirpy/internal/auth"
	"github.com/yuheng-liu/chirpy/internal/database"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	// for converting request json to local struct
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	// for response struct to reply to request
	type response struct {
		User
	}
	// decoding json to struct and handle error
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}
	// hash password and handle error
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}
	// create user and save to db, handle error
	user, err := cfg.DB.CreateUser(params.Email, hashedPassword)
	if err != nil {
		// check if user already exists
		if errors.Is(err, database.ErrAlreadyExists) {
			respondWithError(w, http.StatusConflict, "User already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}
	// all checks passed, send response with proper data
	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}
