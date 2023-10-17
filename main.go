package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)
	router.Mount("/admin", adminRouter)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthzHandler)
	apiRouter.HandleFunc("/reset", apiCfg.resetHandler)
	apiRouter.Get("/chirps", getChirpHandler)
	apiRouter.Post("/chirps", addChirpHandler)
	router.Mount("/api", apiRouter)

	corsMux := middlewareCors(router)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(httpServer.ListenAndServe())
}
