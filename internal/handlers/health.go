package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprint(w, "OK"); err != nil {
		log.Printf("Error writing health response: %v", err)
	}
}
