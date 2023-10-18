package main

import (
	Database "chirpy/internal"
	"encoding/json"
	"log"
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
		Id   int    `json:"id"`
		Body string `json:"body"`
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
	chirp, err = db.CreateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, returnVals{Id: chirp.Id, Body: badWordConvertor(chirp.Body)})
}

func getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	type returnVals []struct {
		Body string `json:"body"`
		Id   int    `json:"id"`
	}

	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})

	var response = returnVals{}
	for _, chirp := range chirps {
		response = append(response, struct {
			Body string `json:"body"`
			Id   int    `json:"id"`
		}{
			Body: chirp.Body,
			Id:   chirp.Id,
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
	id := chi.URLParam(r, "chirpID")
	if err != nil || id == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp id")
		return
	}
	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return

	}
	chirpID, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't convert string to int")
		return
	}
	for _, chirp := range chirps {
		if chirp.Id == chirpID {
			respondWithJSON(w, http.StatusOK, returnVals{Body: chirp.Body, Id: chirp.Id})
			return
		}
	}
	respondWithError(w, http.StatusNotFound, "Chirp not found")
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}
