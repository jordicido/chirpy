package main

import (
	Database "chirpy/internal"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	user, err = db.CreateUser(params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, returnVals{Id: user.Id, Email: user.Email})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	users, err := db.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve users")
		return
	}
	for _, user := range users {
		if user.Email == params.Email {
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password))
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			} else {
				respondWithJSON(w, http.StatusOK, returnVals{Id: user.Id, Email: user.Email})
			}
		}
	}
}
