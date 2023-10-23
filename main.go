package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

var ApiConfig *apiConfig = setUpApiConfig(0, os.Getenv("JWT_SECRET"))

func main() {
	const filepathRoot = "."
	const port = "8080"
	godotenv.Load()

	router := chi.NewRouter()
	fsHandler := ApiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", ApiConfig.metricsHandler)
	router.Mount("/admin", adminRouter)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthzHandler)
	apiRouter.HandleFunc("/reset", ApiConfig.resetHandler)

	apiRouter.Get("/chirps/{chirpID}", getChirpHandler)
	apiRouter.Get("/chirps", getChirpsHandler)
	apiRouter.Post("/chirps", addChirpHandler)

	apiRouter.Post("/users", addUserHandler)
	apiRouter.Put("/users", modifyUserHandler)
	apiRouter.Post("/login", loginHandler)

	router.Mount("/api", apiRouter)

	corsMux := middlewareCors(router)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(httpServer.ListenAndServe())
}
