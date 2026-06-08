package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprint(w, "OK"); err != nil {
		slog.Error("Error writing health response", "error", err)
	}
}
