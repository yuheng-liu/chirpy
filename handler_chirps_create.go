package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	// for converting request json to local struct
	type parameters struct {
		Body string `json:"body"`
	}
	// decoding json to struct and handle error
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}
	// filter out unwanted words and length
	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	// create chirp and save to db, handle error
	chirp, err := cfg.DB.CreateChirp(cleaned)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}
	// all checks passed, send response with proper data
	respondWithJSON(w, http.StatusCreated, Chirp{
		Body: chirp.Body,
		ID:   chirp.ID,
	})
}

func validateChirp(body string) (string, error) {
	// check if body length is too long
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}
	// words to filter out
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	// clean and return result body string
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	// convert body string to a slice of strings
	words := strings.Split(body, " ")
	for i, word := range words {
		// lower case to compare
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			// replcae if matched
			words[i] = "****"
		}
	}
	// convert slice of strings back to single string and return
	cleaned := strings.Join(words, " ")
	return cleaned
}
