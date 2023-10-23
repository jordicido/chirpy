package main

type apiConfig struct {
	fileserverHits int
	jwtScret       string
}

func setUpApiConfig(serverHits int, secretKey string) *apiConfig {
	return &apiConfig{fileserverHits: serverHits, jwtScret: secretKey}
}
