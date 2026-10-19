package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type learner struct {
	Email  string    `json:"email"`
	Course string    `json:"course"`
	Due    time.Time `json:"due"`
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	client := NewInfraiClient(key, http.DefaultClient)
	store := NewDeadlineStore([]learner{{Email: "ada@example.edu", Course: "Data pipelines", Due: time.Now().Add(48 * time.Hour)}})
	http.HandleFunc("/login", loginHandler(client))
	http.HandleFunc("/report", reportHandler(store))
	log.Println("edtech service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loginHandler(client *InfraiClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in struct {
			WidgetRecordID                         string `json:"widget_record_id"`
			Token, Provider, ReturnTo, RedirectURI string
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || in.WidgetRecordID == "" || in.Token == "" || in.Provider == "" {
			http.Error(w, "widget_record_id, token and provider are required", http.StatusBadRequest)
			return
		}
		if err := client.VerifyCaptcha(r.Context(), in.WidgetRecordID, in.Token, "login"); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		url, err := client.AuthorizeURL(r.Context(), in.Provider, in.ReturnTo, in.RedirectURI)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"authorize_url": url})
	}
}

func reportHandler(store *DeadlineStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		json.NewEncoder(w).Encode(store.Report(time.Now()))
	}
}
