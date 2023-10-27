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
		Id          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
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
	respondWithJSON(w, http.StatusCreated, returnVals{Id: user.Id, Email: user.Email, IsChirpyRed: false})
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
		Id          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
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
		user, err = db.UpdateUser(id, params.Email, &params.Password, false)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error updating user")
			return
		}
	}
	respondWithJSON(w, http.StatusOK, returnVals{Id: user.Id, Email: user.Email, IsChirpyRed: false})
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
		IsChirpyRed  bool   `json:"is_chirpy_red"`
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
				respondWithJSON(w, http.StatusOK, returnVals{Id: user.Id, Email: user.Email, Token: accessToken, RefreshToken: refreshToken, IsChirpyRed: user.IsChirpyRed})
			}
		}
	}
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	stringKey := strings.TrimPrefix(r.Header.Get("Authorization"), "ApiKey ")
	if stringKey != ApiConfig.polkaApiKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type parameters struct {
		Data struct {
			UserID int `json:"user_id"`
		} `json:"data"`
		Event string `json:"event"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithoutJSON(w, http.StatusOK)
		return
	}

	user, err := db.GetUser(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve user")
		return
	}
	db.UpdateUser(user.Id, user.Email, nil, true)
	respondWithoutJSON(w, http.StatusOK)
}
