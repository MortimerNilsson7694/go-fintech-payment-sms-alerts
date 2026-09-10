package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	client, err := NewSMSClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		var event PaymentEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		record, err := ProcessPayment(event, client.Send)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	})
	log.Println("payment alerts listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
