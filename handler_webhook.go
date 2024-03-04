package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yuheng-liu/chirpy/internal/auth"
	"github.com/yuheng-liu/chirpy/internal/database"
)

func (cfg *apiConfig) handlerWebhook(w http.ResponseWriter, r *http.Request) {
	// for converting request json to local struct
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		} `json:"data"`
	}
	// retrieve api key from request header
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find api key")
		return
	}
	// check if received apiKey is same as local version
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "API key is invalid")
		return
	}
	// decoding json to struct and handle error
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}
	// check if event is `user.upgraded`, return immediately if not
	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusOK, struct{}{})
		return
	}
	// upgrade user accordingly in the db
	_, err = cfg.DB.UpgradeChirpyRed(params.Data.UserId)
	if err != nil {
		if errors.Is(err, database.ErrNotExist) {
			respondWithError(w, http.StatusNotFound, "Couldn't find user")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user")
		return
	}
	respondWithJSON(w, http.StatusOK, struct{}{})
}
