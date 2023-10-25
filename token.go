package main

import (
	Database "chirpy/internal"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func createToken(id, expiritySeconds int, issuer string) (string, error) {
	issuedAt := time.Now()
	expirity := time.Now().Add(time.Second * time.Duration(expiritySeconds))

	key := []byte(ApiConfig.jwtScret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expirity),
		Subject:   strconv.Itoa(id),
	})
	return token.SignedString(key)
}

func verifyToken(issuer, stringToken string) (int, int) {
	token, err := jwt.ParseWithClaims(stringToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(ApiConfig.jwtScret), nil
	})
	if err != nil {
		return -1, http.StatusUnauthorized
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		if claims.Issuer != issuer {
			return -1, http.StatusUnauthorized
		}
		id, err := strconv.Atoi(claims.Subject)
		if err != nil {
			return -1, http.StatusInternalServerError
		}
		return id, 0
	} else {
		return -1, http.StatusUnauthorized
	}
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	stringToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	id, errorCode := verifyToken("chirpy-refresh", stringToken)
	if id == -1 {
		respondWithError(w, errorCode, "Unauthorized")
		return
	}

	revocations, err := db.GetRevocations()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get revocations")
		return
	}
	for _, revocation := range revocations {
		if revocation.Token == stringToken {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	}
	type returnVals struct {
		Token string `json:"token"`
	}

	accessToken, err := createToken(id, 60*60, "chirpy-access")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating the access-JWT")
		return
	}
	respondWithJSON(w, http.StatusOK, returnVals{Token: accessToken})
}

func revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	db, err := Database.NewDB("")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open database")
		return
	}

	stringToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	id, errorCode := verifyToken("chirpy-refresh", stringToken)
	if id == -1 {
		respondWithError(w, errorCode, "Unauthorized")
		return
	}

	db.RevokeToken(stringToken)
	respondWithoutJSON(w, http.StatusOK)
}
