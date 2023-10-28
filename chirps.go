package main

import (
	Database "chirpy/internal"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func addChirpHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		AuthorId int    `json:"author_id"`
		Id       int    `json:"id"`
		Body     string `json:"body"`
	}

	stringToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	id, errorCode := verifyToken("chirpy-access", stringToken)
	if errorCode != 0 {
		respondWithError(w, errorCode, "Unauthorized")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp too long")
		return
	}
	var chirp Database.Chirp
	chirp, err = db.CreateChirp(params.Body, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, returnVals{AuthorId: chirp.AuthorId, Body: badWordConvertor(chirp.Body), Id: chirp.Id})
}

func getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	var chirps []Database.Chirp
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	stringAuthorId := r.URL.Query().Get("author_id")
	sortMethod := r.URL.Query().Get("sort")

	type returnVals []struct {
		Body     string `json:"body"`
		Id       int    `json:"id"`
		AuthorId int    `json:"author_id"`
	}

	if stringAuthorId != "" {
		authorId, err := strconv.Atoi(stringAuthorId)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
			return
		}
		chirps, err = db.GetChirps(&authorId)
	} else {
		chirps, err = db.GetChirps(nil)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	if sortMethod != "" && sortMethod == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].Id > chirps[j].Id
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].Id < chirps[j].Id
		})
	}

	var response = returnVals{}
	for _, chirp := range chirps {
		response = append(response, struct {
			Body     string `json:"body"`
			Id       int    `json:"id"`
			AuthorId int    `json:"author_id"`
		}{
			Body:     chirp.Body,
			Id:       chirp.Id,
			AuthorId: chirp.AuthorId,
		})
	}
	respondWithJSON(w, http.StatusOK, response)
}

func getChirpHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	type returnVals struct {
		Body string `json:"body"`
		Id   int    `json:"id"`
	}
	stringId := chi.URLParam(r, "chirpID")
	if err != nil || stringId == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp id")
		return
	}

	id, err := strconv.Atoi(stringId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't convert string to int")
		return
	}

	chirp, err := db.GetChirp(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return

	}

	respondWithJSON(w, http.StatusOK, returnVals{Body: chirp.Body, Id: chirp.Id})
}

func deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	stringChirpId := chi.URLParam(r, "chirpID")
	if err != nil || stringChirpId == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp id")
		return
	}
	chirpId, err := strconv.Atoi(stringChirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't convert string to int")
		return
	}

	stringToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	userId, errorCode := verifyToken("chirpy-access", stringToken)
	if errorCode != 0 {
		respondWithError(w, errorCode, "Unauthorized")
	}

	err = db.DeleteChirp(chirpId, userId)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "You can't delete this chirp")
	}
	respondWithoutJSON(w, http.StatusOK)
}

func badWordConvertor(message string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(message)
	for i, word := range words {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}
