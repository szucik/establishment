package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"establishment/v1/internal/database/neo4j"
	"establishment/v1/internal/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	//"establishment/v1/establishment/sqlitle"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	ctx := context.Background()

	var err error
	db, err := database.ConnectToNeo4j(ctx)
	if err != nil {
		log.Fatalf("Error connecting to Neo4j: %v", err)
	}
	defer func(Driver neo4j.DriverWithContext, ctx context.Context) {
		err := Driver.Close(ctx)
		if err != nil {
			log.Fatalf("Error closing driver: %v", err)
		}
	}(db.Driver, ctx)

	//db, err := database.ConnectToSQLite(ctx)
	//if err != nil {
	//	log.Fatalf("Error connecting to SQLite: %v", err)
	//}
	//defer db.Close()

	//if err := database.InitSchema(ctx, db); err != nil {
	//	log.Fatalf("Error initializing SQLite schema: %v", err)
	//}

	http.Handle("/person/", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlePerson(w, r, db)
	})))
	http.Handle("/person", enableCORS(requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlePersonPost(w, r, db)
	}))))
	http.Handle("/relationship", enableCORS(requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRelationship(w, r, db)
	}))))
	http.Handle("/graph", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleGraph(w, r, db)
	})))
	http.Handle("/persons", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlePersons(w, r, db)
	})))
	http.Handle("/register", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRegister(w, r, db)
	})))
	http.Handle("/login", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleLogin(w, r, db)
	})))
	http.Handle("/logout", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleLogout(w, r, db)
	})))
	http.Handle("/check-session", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleCheckSession(w, r, db)
	})))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	log.Println("Server started on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlePerson(w http.ResponseWriter, r *http.Request, db models.Database) {
	if r.Method != http.MethodGet {
		log.Printf("Unsupported method %s for /person (GET)", r.Method)
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := strings.TrimPrefix(r.URL.Path, "/person/")
	if id == "" {
		log.Printf("No person ID provided in request: %s", r.URL.Path)
		http.Error(w, "Person ID required", http.StatusBadRequest)
		return
	}

	person, err := db.GetPerson(ctx, id)
	if errors.Is(err, database.ErrNoSuchPerson) {
		log.Printf("Person not found for ID: %s", id)
		http.Error(w, "Person not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error fetching person for ID %s: %v", id, err)
		http.Error(w, "Error fetching person: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, person)
}

func handlePersonPost(w http.ResponseWriter, r *http.Request, db models.Database) {
	if r.Method != http.MethodPost {
		log.Printf("Unsupported method %s for /person (POST)", r.Method)
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var person models.Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		log.Printf("Invalid input data in POST /person: %v", err)
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	if person.ID == "" || person.Name == "" {
		log.Printf("Missing required fields in POST /person: %+v", person)
		http.Error(w, "ID and name are required", http.StatusBadRequest)
		return
	}

	if err := db.AddPerson(ctx, person); err != nil {
		log.Printf("Failed to add person: %v", err)
		http.Error(w, "Failed to add person: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleRelationship(w http.ResponseWriter, r *http.Request, db models.Database) {
	if r.Method != http.MethodPost {
		log.Printf("Unsupported method %s for /relationship", r.Method)
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var rel models.Relationship
	if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
		log.Printf("Invalid input data in POST /relationship: %v", err)
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.AddRelationship(ctx, rel); err != nil {
		log.Printf("Failed to add relationship: %v", err)
		http.Error(w, "Failed to add relationship: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleGraph(w http.ResponseWriter, r *http.Request, db models.Database) {
	if r.Method != http.MethodGet {
		log.Printf("Unsupported method %s for /graph", r.Method)
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	graph, err := db.GetGraph(ctx)
	if err != nil {
		log.Printf("Error fetching graph: %v", err)
		http.Error(w, "Error fetching graph: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, graph)
}

func handlePersons(w http.ResponseWriter, r *http.Request, db models.Database) {
	if r.Method != http.MethodGet {
		log.Printf("Unsupported method %s for /persons", r.Method)
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	persons, err := db.GetPersons(ctx)
	if err != nil {
		log.Printf("Error fetching persons: %v", err)
		http.Error(w, "Error fetching persons: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, persons)
}

func handleRegister(w http.ResponseWriter, r *http.Request, db models.Database) {
	if r.Method != http.MethodPost {
		log.Printf("Unsupported method %s for /register", r.Method)
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("Invalid input data in POST /register: %v", err)
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	if input.Login == "" || input.Email == "" || input.Password == "" {
		log.Printf("Missing required fields in POST /register")
		http.Error(w, "Login, email, and password are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Error hashing password: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:       uuid.New().String(),
		Login:    input.Login,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	if err := db.AddUser(ctx, user); err != nil {
		if err == database.ErrUserExists {
			log.Printf("User already exists: login=%s, email=%s", input.Login, input.Email)
			http.Error(w, "User with this login or email already exists", http.StatusBadRequest)
			return
		}
		log.Printf("Failed to register user: %v", err)
		http.Error(w, "Failed to register user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleLogin(w http.ResponseWriter, r *http.Request, db models.Database) {
	if r.Method != http.MethodPost {
		log.Printf("Unsupported method %s for /login", r.Method)
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("Invalid input data in POST /login: %v", err)
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := db.GetUserByLogin(ctx, input.Login)
	if err != nil {
		log.Printf("Invalid login: %s", input.Login)
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		log.Printf("Invalid password for login: %s", input.Login)
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	session := models.Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	if err := db.CreateSession(ctx, session); err != nil {
		log.Printf("Error creating session for user %s: %v", user.ID, err)
		http.Error(w, "Error creating session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Session created: session_id=%s, user_id=%s, expires_at=%d", session.ID, user.ID, session.ExpiresAt)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	log.Printf("User logged in: %s, cookie set with session_id=%s", user.Login, session.ID)
	w.WriteHeader(http.StatusOK)
}

func handleLogout(w http.ResponseWriter, r *http.Request, db models.Database) {
	if r.Method != http.MethodPost {
		log.Printf("Unsupported method %s for /logout", r.Method)
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessionID, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session_id cookie in /logout")
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := db.DeleteSession(ctx, sessionID.Value); err != nil {
		log.Printf("Error logging out for session %s: %v", sessionID.Value, err)
		http.Error(w, "Error logging out: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	log.Printf("User logged out, session: %s", sessionID.Value)
	w.WriteHeader(http.StatusOK)
}

func handleCheckSession(w http.ResponseWriter, r *http.Request, db models.Database) {
	if r.Method != http.MethodGet {
		log.Printf("Unsupported method %s for /check-session", r.Method)
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessionID, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session_id cookie in /check-session")
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	log.Printf("Checking session: session_id=%s", sessionID.Value)

	session, err := db.GetSession(ctx, sessionID.Value)
	if err != nil || session.ExpiresAt < time.Now().Unix() {
		log.Printf("Session inactive or expired: session_id=%s, error: %v", sessionID.Value, err)
		http.Error(w, "Session inactive or expired", http.StatusUnauthorized)
		return
	}

	user, err := db.GetUserByID(ctx, session.UserID)
	if err != nil {
		log.Printf("User not found for session %s: %v", sessionID.Value, err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	log.Printf("Session valid: session_id=%s, user=%s", sessionID.Value, user.Login)

	writeJSON(w, struct {
		Login string `json:"login"`
	}{Login: user.Login})
}

// requireAuth middleware
func requireAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db, ok := r.Context().Value("db").(models.Database)
		if !ok {
			log.Printf("No database in context for request: %s", r.URL.Path)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		sessionID, err := r.Cookie("session_id")
		if err != nil {
			log.Printf("No session_id cookie in request: %s", r.URL.Path)
			http.Error(w, "Login required", http.StatusUnauthorized)
			return
		}

		session, err := db.GetSession(ctx, sessionID.Value)
		if err != nil || session.ExpiresAt < time.Now().Unix() {
			log.Printf("Session inactive or expired for ID %s: %v", sessionID.Value, err)
			http.Error(w, "Session inactive or expired", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, "db", db)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:5500",
			"http://127.0.0.1:5500",
		}

		allowOrigin := ""
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				allowOrigin = origin
				break
			}
		}

		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		log.Printf("CORS: Origin=%s, AllowOrigin=%s", origin, allowOrigin)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		db, ok := r.Context().Value("db").(models.Database)
		if ok {
			ctx := context.WithValue(r.Context(), "db", db)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
