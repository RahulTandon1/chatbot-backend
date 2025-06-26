// main entry point that start the server
package main

// TODO: figure out what go.sum does

import (
	"log"
	"net/http"
	"os"
	"github.com/joho/godotenv"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	// if r.URL.Path != "/" {
	// 	http.Error(w, "Not found", http.StatusNotFound)
	// 	return
	// }
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ------- start of liberal permissions ------
		// CORS headers
		
		// for PROD: w.Header().Set("Access-Control-Allow-Origin", "https://your-frontend-url.com")
		w.Header().Set("Access-Control-Allow-Origin", "*") // or restrict to "http://localhost:3000"
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// ------- end of liberal permissions ------

		handler(w, r)
	}
}

func main() {
	// load .env file (local dev); 
	// if not present, don't panic. OS might already have it in prod.
	_ = godotenv.Load()
	// err != nil {
	// 	log.Fatal("Error loading .env file")
	// }
	// log.Println(".env loaded successfully")

	http.HandleFunc("/", serveHome)

	// Route for websocket
	http.HandleFunc("/ws", withCORS(handleWebSocket)) // handlWebSocket defined in another file
	http.HandleFunc("/history", withCORS(handleHistory)) // defined in another file

	// TODO: also read port from .env file if not in OS environment vars
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// TODO: figure out if Go has template string/literals
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Printf("Server started on http://localhost:%s", port)
	// log.Fatal(http.ListenAndServe(":"+port, nil))

}