package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/assets/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", healthzHandler)

	corsMux := middlewareCors(mux)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: corsMux,
	}

	err := httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
