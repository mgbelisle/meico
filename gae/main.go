package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	// Load configs
	configs := struct {
		StripeSecretKey string `json:"stripeSecretKey"`
	}{}
	configFile, err := os.Open("configs/prod.json")
	if err != nil {
		panic(err)
	}
	if err := json.NewDecoder(configFile).Decode(&configs); err != nil {
		panic(err)
	}
	configFile.Close()

	// Allow all origins
	cors := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			switch r.Method {
			case "OPTIONS":
				w.Header().Set("Access-Control-Allow-Methods", r.Header.Get("Access-Control-Request-Method"))
				w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
				return
			default:
				h(w, r)
				return
			}
		}
	}

	// Create a stripe charge
	http.HandleFunc("/stripe/v1/payment_intents", cors(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			stripeReq, _ := http.NewRequest("POST", "https://api.stripe.com/v1/payment_intents", r.Body)
			stripeReq.Header.Set("Authorization", "Bearer "+configs.StripeSecretKey)
			if contentType := r.Header.Get("Content-Type"); contentType != "" {
				stripeReq.Header.Set("Content-Type", contentType)
			}
			stripeResp, err := http.DefaultClient.Do(stripeReq)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(stripeResp.StatusCode)
			io.Copy(w, stripeResp.Body)
			return
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
