package main

import (
	Database "chirpy/internal"
	"encoding/json"
	"net/http"
)

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	type parameters struct {
		Email string `json:"email"`
	}
	type returnVals struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	var user Database.User
	user, err = db.CreateUser(params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, returnVals{Id: user.Id, Email: user.Email})
}
