#!/bin/bash
# Użycie: API_LOGIN=... API_PASSWORD=... ./add_person.sh
# Dane osób i relacji dopisz na końcu pliku wywołaniami add_person / add_relationship.

set -euo pipefail

API_URL="${API_URL:-http://localhost:8080}"
API_LOGIN="${API_LOGIN:-test}"
API_PASSWORD="${API_PASSWORD:-test}"

COOKIE_JAR="$(mktemp)"
RESPONSE_BODY="$(mktemp)"
trap 'rm -f "$COOKIE_JAR" "$RESPONSE_BODY"' EXIT

post_json() {
  local path="$1"
  local json="$2"
  curl -s -b "$COOKIE_JAR" -c "$COOKIE_JAR" -o "$RESPONSE_BODY" -w "%{http_code}" \
    -X POST "$API_URL$path" -H "Content-Type: application/json" -d "$json"
}

fail_with_response() {
  echo "$1 (HTTP $2): $(cat "$RESPONSE_BODY")" >&2
  exit 1
}

login() {
  local status
  status=$(post_json /login "{\"login\": \"$API_LOGIN\", \"password\": \"$API_PASSWORD\"}")
  [ "$status" = "200" ] || fail_with_response "Błąd logowania" "$status"
  echo "Zalogowano jako $API_LOGIN"
}

# add_person '<json osoby>'  np. '{"id":"jan-kowalski","name":"Jan Kowalski","occupation":"poseł"}'
add_person() {
  local status
  status=$(post_json /person "$1")
  [ "$status" = "201" ] || fail_with_response "Błąd dodawania osoby $1" "$status"
  echo "Dodano osobę: $1"
}

# add_relationship <source_id> <target_id> <typ> [szczegóły]
add_relationship() {
  local status
  status=$(post_json /relationship \
    "{\"source_id\": \"$1\", \"target_id\": \"$2\", \"type\": \"$3\", \"details\": \"${4:-}\"}")
  [ "$status" = "201" ] || fail_with_response "Błąd dodawania relacji $1 -> $2 ($3)" "$status"
  echo "Dodano relację: $1 -> $2 ($3)"
}

login

echo "Dodawanie nowych osób..."
