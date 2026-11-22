# Elity Polski — układ zamknięty

## Acknowledgments

This project was developed with the assistance of **Grok**, an AI tool created by xAI, which provided support in writing, debugging, and optimizing code, as well as generating documentation. Later refactoring and the frontend redesign were done with **Claude Code** (Anthropic). Special thanks to the open-source community for providing libraries like `go-sqlite3`, `neo4j-go-driver`, Vue and d3.

## Overview

A web application for mapping connections between people: add persons, connect them with typed relationships (family, party, business, …) and explore the resulting network as a graph drawn on an animated galaxy background.

- **Backend:** Go, hexagonal architecture, SQLite (default) or Neo4j.
- **Frontend:** static HTML pages with Vue 3 and d3, served by the same Go server.

## Features

- **Graph of connections** (`/static/`): every person is a planet, every relationship a beam of light labelled with its type. Drag nodes, hover for details, click to open a profile.
- **Person profile** (`/static/person_view.html?id=<id>`): photo (or a planet with initials), occupation, description, link to X, and a graph of only that person's connections with the person pinned in the centre.
- **Search** (top right on the graph, profile and panel pages): finds people by name, occupation or description; ignores Polish diacritics and matches surname gender forms (*Zieliński* / *Zielińska*). Arrow keys + Enter, Esc closes.
- **Panel** (`/static/backend.html`, login required): add persons and relationships. The relationship type is free text with suggestions — type a new one (e.g. `pracownik`) to create it.
- **Accounts:** register, log in (24 h session cookie), log out.

Rules enforced by the domain:

- A person needs an ID and a name; IDs are unique.
- A relationship needs both ends, a type, and cannot connect a person with themselves; both persons must exist.
- A relationship is unique per *(source, target, type)*; saving it again updates its details.

## Project Structure

```
establishment/
├── main.go, config.go        # Composition root: env config, wiring, HTTP server with timeouts and graceful shutdown
├── internal/
│   ├── person/               # Domain: Person, Relationship, Graph, validation, Service + Repository port
│   ├── auth/                 # Domain: User, Session, Registration, Service + repository/hasher ports
│   ├── httpapi/              # HTTP adapter: routes, JSON DTOs, session & CORS middleware, error mapping
│   ├── sqlitestore/          # SQLite adapter (schema migrations via PRAGMA user_version)
│   ├── neo4jstore/           # Neo4j adapter (uniqueness constraints created on start)
│   └── password/             # bcrypt adapter for auth.PasswordHasher
├── static/
│   ├── index.html            # Graph of all connections
│   ├── person_view.html      # Person profile with their own connections graph
│   ├── backend.html          # Panel: add persons and relationships
│   ├── login.html            # Login and registration
│   ├── graph.js              # Shared graph renderer (full graph and "person in the centre" mode)
│   ├── search.js             # Person search box (mounted on [data-person-search])
│   ├── cosmos.js             # Animated galaxy background (canvas)
│   └── theme.css             # Shared dark/cosmic theme
└── add_person.sh             # Script for bulk-adding persons and relationships via the API
```

Dependencies point inwards: `httpapi`, `sqlitestore`, `neo4jstore` → `person`, `auth`. Domain packages import nothing from infrastructure and define the interfaces they need (consumer-side ports).

## Requirements

- **Go** 1.23+ and a C compiler (`gcc`) — `go-sqlite3` uses cgo.
- **Neo4j 5.x** — only when running with `STORAGE=neo4j`.

## Running Locally

```bash
go run .                      # SQLite in ./data.db, http://localhost:8080
ADDR=:8081 go run .           # another port
STORAGE=neo4j NEO4J_PASSWORD=... go run .
```

Open `http://localhost:8080/` (redirects to the graph). The frontend uses relative API URLs, so it works on whatever address the server runs.

| Variable | Default | Meaning |
|---|---|---|
| `ADDR` | `:8080` | Listen address |
| `STORAGE` | `sqlite` | `sqlite` or `neo4j` |
| `SQLITE_PATH` | `./data.db` | SQLite file (created and migrated automatically) |
| `NEO4J_URI` / `NEO4J_USER` / `NEO4J_PASSWORD` | `neo4j://localhost:7687` / `neo4j` / required | Neo4j connection |
| `ALLOWED_ORIGINS` | `http://localhost:5500,http://127.0.0.1:5500` | CORS origins (comma-separated), only needed when the frontend is served from elsewhere |
| `COOKIE_SECURE` | `false` | Set `true` when served over HTTPS |
| `STATIC_DIR` | `static` | Directory served at `/static/` |

Local Neo4j:

```bash
docker run -p 7687:7687 -p 7474:7474 -e NEO4J_AUTH=neo4j/<password> neo4j
```

### Adding data from a script

`add_person.sh` logs in and exposes `add_person '<json>'` and `add_relationship <source> <target> <type> [details]`. Append calls at the end of the file and run:

```bash
API_URL=http://localhost:8080 API_LOGIN=<login> API_PASSWORD=<password> ./add_person.sh
```

## API Endpoints

| Endpoint | Method | Description | Request Body |
|----------|--------|-------------|--------------|
| `/graph` | GET | All persons (`nodes`) and relationships (`edges`) | — |
| `/persons` | GET | All persons | — |
| `/person/{id}` | GET | One person | — |
| `/person` | POST | Add a person 🔒 | `{"id","name","occupation","image_url","twitter","description"}` |
| `/relationship` | POST | Add or update a relationship 🔒 | `{"source_id","target_id","type","details"}` |
| `/register` | POST | Register a user | `{"login","email","password"}` |
| `/login` | POST | Log in, sets `session_id` cookie | `{"login","password"}` |
| `/logout` | POST | Log out | — (session cookie) |
| `/check-session` | GET | Logged-in user's login 🔒 | — |

🔒 requires the session cookie. Errors: `400` invalid input, `401` not logged in / bad credentials, `404` unknown person, `409` duplicate person or user, `500` internal error (details only in server logs). Request bodies are limited to 1 MB.

## Database Switching

Set `STORAGE=sqlite` or `STORAGE=neo4j`. Both adapters implement the same ports (`person.Repository`, `auth.UserRepository`, `auth.SessionRepository`) with the same behaviour, so the rest of the application does not change.

## Testing

```bash
go vet ./...
go test -race ./...
```

Unit tests cover the domain (`person`, `auth`) and the HTTP API (`httpapi`) through their public functions, using fakes — no database or network needed.

## Deployment

The Go server serves both the API and the frontend, so deploy it as a single service:

- Persist the SQLite file on a volume and set `SQLITE_PATH` (e.g. `/data/data.db`), or point it at Neo4j with `STORAGE=neo4j` and the `NEO4J_*` variables.
- Serve over HTTPS and set `COOKIE_SECURE=true`.

## Known Limitations

- There is no editing or deleting of persons and relationships through the API yet.
- Expired sessions are rejected but not removed from the database.
- The Neo4j adapter has not been tested against a live instance after the refactor.

## License

MIT License (see `LICENSE` file for details).
