package main

import (
	Database "chirpy/internal"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func verifyToken(stringToken string) (int, int) {
	token, err := jwt.ParseWithClaims(stringToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(ApiConfig.jwtScret), nil
	})
	if err != nil {
		return -1, http.StatusUnauthorized
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		id, err := strconv.Atoi(claims.Subject)
		if err != nil {
			return -1, http.StatusInternalServerError
		}
		return id, 0
	} else {
		return -1, http.StatusUnauthorized
	}
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

	id, errorCode := verifyToken(stringToken)
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
		Id    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
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
				var expiritySeconds time.Time
				issuedAt := time.Now()
				if params.ExpiresInSeconds == nil {
					expiritySeconds = time.Now().AddDate(0, 0, 1)
				} else {
					expiritySeconds = time.Now().Add(time.Second * time.Duration(*params.ExpiresInSeconds))
				}
				key := []byte(ApiConfig.jwtScret)
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Issuer:    "chirpy",
					IssuedAt:  jwt.NewNumericDate(issuedAt),
					ExpiresAt: jwt.NewNumericDate(expiritySeconds),
					Subject:   strconv.Itoa(user.Id),
				})
				stringToken, err := token.SignedString(key)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Error creating the JWT")
					return
				}
				respondWithJSON(w, http.StatusOK, returnVals{Id: user.Id, Email: user.Email, Token: stringToken})
			}
		}
	}
}
