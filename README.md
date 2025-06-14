# Establishment - Graph Application 

## Acknowledgments

This project was developed with the assistance of **Grok**, an AI tool created by xAI, which provided support in writing, debugging, and optimizing code, as well as generating documentation. Special thanks to the open-source community for providing libraries like `go-sqlite3` and `neo4j-go-driver`.


## Overview

The **Establishment Graph Application** is a web-based platform for managing and visualizing relationships between individuals. It allows users to create profiles for persons, define relationships between them, and view the resulting network as a graph. The application supports user authentication, session management, and persistent storage using either SQLite or Neo4j databases. The backend is built in Go, and the frontend uses static HTML with JavaScript for dynamic interactions.

This project was developed with the assistance of artificial intelligence tools to streamline coding, debugging, and documentation processes.

## Features

- **Person Management**:
    - Add persons with attributes: ID, name, occupation, image URL, Twitter handle, and description.
    - Retrieve individual person details by ID or list all persons.
- **Relationship Management**:
    - Create directed relationships between persons with a type and details.
    - Prevent self-referential relationships and validate person existence.
- **Graph Visualization**:
    - Fetch the entire graph of persons (nodes) and relationships (edges) for visualization.
- **User Authentication**:
    - Register users with unique login and email.
    - Login to create sessions with expiration times.
    - Logout to delete sessions.
- **Database Flexibility**:
    - Supports SQLite for lightweight, file-based storage.
    - Supports Neo4j for graph-based storage, optimized for relationship queries.
- **REST API**:
    - Endpoints for managing persons, relationships, users, sessions, and the graph.
- **Frontend**:
    - Static HTML pages for viewing person details and the graph.
    - JavaScript for fetching data from the backend API.

## Project Structure

```
establishment/
├── database/
│   ├── neo4j.go        # Neo4j database implementation
│   ├── sqlite3.go      # SQLite database implementation
├── models/
│   ├── models.go       # Data structures for Person, Relationship, User, Session, Graph
├── static/
│   ├── person_view.html # Frontend page for person details
│   ├── graph_view.html  # Frontend page for graph visualization
│   ├── *.js            # JavaScript for frontend interactivity
├── main.go             # Main application entry point with HTTP server and handlers
├── go.mod              # Go module dependencies
├── fly.toml            # Fly.io deployment configuration
└── README.md           # Project documentation (this file)
```

## Requirements

- **Go**: Version 1.18 or later
- **Databases** (one of):
    - **SQLite**: No additional setup required (uses `data.db` file).
    - **Neo4j**: Version 5.x, accessible via `neo4j://` URI (e.g., local or Neo4j Aura).
- **Node.js** (for frontend testing): Version 14.x or later
- **Docker** (optional): For running Neo4j locally.
- **Dependencies**:
    - Go: `github.com/mattn/go-sqlite3`, `github.com/neo4j/neo4j-go-driver/v5`
    - Node.js: `live-server` for local frontend testing

## Installation

1. **Clone the Repository**:
   ```bash
   git clone <repository-url>
   cd establishment
   ```

2. **Install Go Dependencies**:
   ```bash
   go mod tidy
   ```

3. **Set Up Database**:
    - **SQLite**:
        - No setup needed; `data.db` will be created automatically.
    - **Neo4j** (Local):
      ```bash
      docker run -p 7687:7687 -p 7474:7474 -e NEO4J_AUTH=neo4j/secretgraph neo4j
      ```
      Set environment variables:
      ```bash
      export NEO4J_URI=neo4j://localhost:7687
      export NEO4J_USER=neo4j
      export NEO4J_PASSWORD=secretgraph
      ```
    - **Neo4j (Cloud)**:
        - Use Neo4j Aura or a hosted instance.
        - Set environment variables with provided credentials.

4. **Install Frontend Dependencies** (for local testing):
   ```bash
   npm install -g live-server
   ```

## Running Locally

1. **Start the Backend**:
    - For SQLite:
      ```bash
      go run main.go
      ```
    - For Neo4j:
      Ensure Neo4j is running and environment variables are set, then:
      ```bash
      go run main.go
      ```
    - The server runs on `http://localhost:8080`.

2. **Start the Frontend**:
   ```bash
   npx live-server --port=5500
   ```
    - Access:
        - Person view: `http://localhost:5500/static/person_view.html?id=<uuid>`
        - Graph view: `http://localhost:5500/static/graph_view.html`

3. **Test API Endpoints**:
    - Get graph:
      ```bash
      curl http://localhost:8080/graph
      ```
    - Register user:
      ```bash
      curl -X POST http://localhost:8080/register -d '{"login":"test","email":"test@example.com","password":"password"}'
      ```
    - Login:
      ```bash
      curl -X POST http://localhost:8080/login -d '{"login":"test","password":"password"}'
      ```

## API Endpoints

| Endpoint | Method | Description | Request Body |
|----------|--------|-------------|--------------|
| `/graph` | GET | Retrieve the entire graph | None |
| `/person` | POST | Add a person | `{"id":"uuid","name":"string","occupation":"string","image_url":"string","twitter":"string","description":"string"}` |
| `/person/{id}` | GET | Get person by ID | None |
| `/relationship` | POST | Add a relationship | `{"source_id":"uuid","target_id":"uuid","type":"string","details":"string"}` |
| `/register` | POST | Register a user | `{"login":"string","email":"string","password":"string"}` |
| `/login` | POST | Login and create session | `{"login":"string","password":"string"}` |
| `/logout` | POST | Logout and delete session | None (requires session cookie) |

## Database Switching

The application supports both SQLite and Neo4j via a `Database` interface. To switch databases, modify `main.go`:

- **SQLite**:
  ```go
  db, err := database.ConnectToSQLite(ctx)
  if err != nil {
      log.Fatalf("Error connecting to SQLite: %v", err)
  }
  defer db.Close()
  if err := database.InitSchema(ctx, db); err != nil {
      log.Fatalf("Error initializing schema: %v", err)
  }
  ```

- **Neo4j**:
  ```go
  db, err := database.ConnectToNeo4j(ctx)
  if err != nil {
      log.Fatalf("Error connecting to Neo4j: %v", err)
  }
  defer db.Driver.Close(ctx)
  ```

## Deployment

### Backend (Fly.io)

1. **Install Flyctl**:
   ```bash
   curl -L https://fly.io/install.sh | sh
   ```

2. **Create Fly.io App**:
   ```bash
   flyctl launch
   ```

3. **Configure `fly.toml`**:
   Ensure `fly.toml` is set up (see previous artifact_id: `e015723b-1a8f-4bdd-b972-85efd6eaaf26`).

4. **Set Environment Variables** (for Neo4j):
   ```bash
   flyctl secrets set NEO4J_URI=neo4j://your-neo4j-host
   flyctl secrets set NEO4J_USER=neo4j
   flyctl secrets set NEO4J_PASSWORD=your-password
   ```

5. **Deploy**:
   ```bash
   flyctl deploy
   ```

6. **SQLite Deployment**:
    - Use a persistent volume:
      ```bash
      flyctl volumes create data --region <region>
      ```
      Update `fly.toml`:
      ```toml
      [[mounts]]
        source = "data"
        destination = "/data"
      ```
      Modify `database/sqlite3.go` to use `/data/data.db`.

### Frontend (Netlify)

1. **Create Netlify Site**:
    - Drag and drop the `static/` folder to Netlify's dashboard or use:
      ```bash
      npm install -g netlify-cli
      netlify deploy
      ```

2. **Configure**:
    - Set build directory to `static/`.
    - Update API URLs in `static/*.js` to point to the Fly.io backend (e.g., `https://your-app.fly.dev`).

3. **Deploy**:
   ```bash
   netlify deploy --prod
   ```

## Testing

- **Backend**:
  ```bash
  go test ./...
  ```
  (Add tests to `database/` and handlers if needed.)

- **API**:
  Use `curl` or Postman to test endpoints (see "Test API Endpoints" above).

- **Database**:
    - SQLite: Open `data.db` with `sqlite3 data.db` and run:
      ```sql
      .schema
      SELECT * FROM Person;
      ```
    - Neo4j: Use the Neo4j Browser (`http://localhost:7474`) and run:
      ```cypher
      MATCH (n) RETURN n LIMIT 10
      ```

- **Frontend**:
  Verify pages load correctly and data is fetched from the API.

## Troubleshooting

- **Backend Errors**:
    - Check logs: `go run main.go` or `flyctl logs`.
    - Ensure database is accessible (Neo4j URI or SQLite file permissions).
- **Database Issues**:
    - SQLite: Verify `data.db` is writable.
    - Neo4j: Check credentials and connectivity (`telnet localhost 7687`).
- **Frontend Issues**:
    - Open browser console (F12) to check for JavaScript errors.
    - Verify API URLs in `static/*.js`.
- **Dependencies**:
  ```bash
  go mod tidy
  npm install
  ```

## Contributing

1. Fork the repository.
2. Create a feature branch: `git checkout -b feature-name`.
3. Commit changes: `git commit -m "Add feature"`.
4. Push to branch: `git push origin feature-name`.
5. Open a pull request.

## License

MIT License (see `LICENSE` file for details).