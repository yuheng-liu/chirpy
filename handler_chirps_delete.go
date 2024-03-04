package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yuheng-liu/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	// retrieve the argument parameter, in this case the chirpID
	chirpID, err := strconv.Atoi(chi.URLParam(r, "chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}
	// retrieve the jwt from request header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	// check if retrieved jwt is valid, get back user ID as subject value
	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}
	// convert user ID to int
	userID, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse user ID")
		return
	}
	// get chirp based on the chirpID
	dbChirp, err := cfg.DB.GetChirp(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp")
		return
	}
	// check if user id of request matches user id of chirp
	if dbChirp.AuthorID != userID {
		respondWithError(w, http.StatusForbidden, "You can't delete this chirp")
		return
	}
	// try to delete chirp from db and handle error if exists
	err = cfg.DB.DeleteChirp(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp")
		return
	}
	// all checks passed, send response with proper data
	respondWithJSON(w, http.StatusOK, struct{}{})
}
