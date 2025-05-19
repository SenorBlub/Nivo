package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var ChiselIp = os.Getenv("CHISEL_IP")
var GlintIp = os.Getenv("GLINT_IP")
var ResonoIp = os.Getenv("RESONO_IP")
var httpPort = os.Getenv("PORT")

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Allow specific headers and methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func main() {
	http.HandleFunc("/view", enableCORS(viewHandler))
	http.HandleFunc("/listen", enableCORS(listenHandler))
	http.HandleFunc("/read", enableCORS(readHandler))
	http.HandleFunc("/document", enableCORS(documentHandler))
	http.HandleFunc("/think", enableCORS(thinkHandler))
	http.HandleFunc("/explain", enableCORS(explainHandler))
	http.HandleFunc("/remember", enableCORS(rememberHandler))
	http.HandleFunc("/ask", enableCORS(askHandler))

	fmt.Println("🤖 Nivo API running on port " + httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}
