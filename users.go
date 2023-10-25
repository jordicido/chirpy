package main

import (
	Database "chirpy/internal"
	"encoding/json"
	"net/http"
	"strings"

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

func modifyUserHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	var user Database.User
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

	stringToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	id, errorCode := verifyToken("chirpy-access", stringToken)
	if id == -1 {
		respondWithError(w, errorCode, "Unauthorized")
		return
	} else {
		user, err = db.UpdateUser(id, params.Email, params.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error updating user")
			return
		}
	}
	respondWithJSON(w, http.StatusOK, returnVals{Id: user.Id, Email: user.Email})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}
	type returnVals struct {
		Id           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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
				return
			} else {
				accessToken, err := createToken(user.Id, 60*60, "chirpy-access")
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Error creating the access-JWT")
					return
				}
				refreshToken, err := createToken(user.Id, 60*60*24*60, "chirpy-refresh")
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Error creating the refresh-JWT")
					return
				}
				respondWithJSON(w, http.StatusOK, returnVals{Id: user.Id, Email: user.Email, Token: accessToken, RefreshToken: refreshToken})
			}
		}
	}
}
