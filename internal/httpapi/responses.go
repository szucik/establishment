package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"establishment/v1/internal/auth"
	"establishment/v1/internal/person"
)

const maxRequestBodyBytes = 1 << 20

var statusForDomainError = []struct {
	err    error
	status int
}{
	{person.ErrMissingID, http.StatusBadRequest},
	{person.ErrMissingName, http.StatusBadRequest},
	{person.ErrMissingRelationshipEnd, http.StatusBadRequest},
	{person.ErrMissingRelationshipType, http.StatusBadRequest},
	{person.ErrSelfRelationship, http.StatusBadRequest},
	{auth.ErrMissingCredentials, http.StatusBadRequest},
	{person.ErrNotFound, http.StatusNotFound},
	{person.ErrAlreadyExists, http.StatusConflict},
	{auth.ErrUserAlreadyExists, http.StatusConflict},
	{auth.ErrInvalidCredentials, http.StatusUnauthorized},
	{auth.ErrSessionNotFound, http.StatusUnauthorized},
	{auth.ErrSessionExpired, http.StatusUnauthorized},
	{auth.ErrUserNotFound, http.StatusUnauthorized},
}

// writeError is the single place where request errors are handled:
// domain errors become client responses, anything else is logged and hidden.
func writeError(w http.ResponseWriter, err error) {
	for _, known := range statusForDomainError {
		if errors.Is(err, known.err) {
			http.Error(w, known.err.Error(), known.status)
			return
		}
	}
	log.Printf("internal error: %v", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	encoded, err := json.Marshal(body)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(encoded); err != nil {
		log.Printf("writing response: %v", err)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	body := http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	if err := json.NewDecoder(body).Decode(destination); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return false
	}
	return true
}
