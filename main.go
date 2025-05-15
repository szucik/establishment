package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"establishment/v1/establishment/models"
	database "establishment/v1/establishment/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var driver neo4j.DriverWithContext

func main() {
	ctx := context.Background()

	var err error
	driver, err = database.ConnectToNeo4j(ctx)
	if err != nil {
		log.Fatalf("Błąd połączenia z Neo4j: %v", err)
	}
	defer driver.Close(ctx)

	http.Handle("/person", enableCORS(http.HandlerFunc(handlePerson)))
	http.Handle("/relationship", enableCORS(http.HandlerFunc(handleRelationship)))
	http.Handle("/graph", enableCORS(http.HandlerFunc(handleGraph)))
	http.Handle("/persons", enableCORS(http.HandlerFunc(handlePersons)))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	log.Println("Serwer uruchomiony na porcie :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// POST /person
func handlePerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Tylko POST", http.StatusMethodNotAllowed)
		return
	}

	var person models.Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		http.Error(w, "Błędne dane wejściowe", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.AddPerson(ctx, driver, person); err != nil {
		http.Error(w, "Nie udało się dodać osoby: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// POST /relationship
func handleRelationship(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Tylko POST", http.StatusMethodNotAllowed)
		return
	}

	var rel models.Relationship
	if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
		http.Error(w, "Błędne dane wejściowe", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.AddRelationship(ctx, driver, rel); err != nil {
		http.Error(w, "Nie udało się dodać relacji: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /graph
func handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Tylko GET", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	graph, err := database.GetGraph(ctx, driver)
	if err != nil {
		http.Error(w, "Błąd pobierania grafu: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, graph)
}

// GET /persons
func handlePersons(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Tylko GET", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	persons, err := database.GetPersons(ctx, driver)
	if err != nil {
		http.Error(w, "Błąd pobierania osób: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, persons)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Błąd kodowania JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
