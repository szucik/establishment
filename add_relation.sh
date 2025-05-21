#!/bin/bash

# Skrypt do dodawania przykładowych relacji między osobami w bazie Neo4j przez API

# Funkcja do sprawdzania odpowiedzi curl
check_status() {
    if [ $1 -ne 0 ]; then
        echo "Błąd: Polecenie curl zwróciło kod błędu $1"
        exit 1
    fi
}

echo "Dodawanie relacji między osobami..."

# Dodawanie relacji 1: Ewa Malinowska i Jakub Szymański (COLLEAGUE)
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/relationship -H "Content-Type: application/json" -d '{
    "source_id": "m1",
    "target_id": "m2",
    "type": "COLLEAGUE",
    "details": "Współpraca w komisji transportu"
}'
check_status $?
echo "Dodano relację między m1 a m2"

# Dodawanie relacji 2: Natalia Kowalczyk i Monika Zielińska (COLLEAGUE)
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/relationship -H "Content-Type: application/json" -d '{
    "source_id": "m3",
    "target_id": "m5",
    "type": "COLLEAGUE",
    "details": "Współpraca w redakcji portalu"
}'
check_status $?
echo "Dodano relację między m3 a m5"

# Dodawanie relacji 3: Grzegorz Wiśniewski i Bartosz Kaczmarek (COLLEAGUE)
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/relationship -H "Content-Type: application/json" -d '{
    "source_id": "m4",
    "target_id": "m6",
    "type": "COLLEAGUE",
    "details": "Wspólne projekty w sektorze energetycznym"
}'
check_status $?
echo "Dodano relację między m4 a m6"

# Dodawanie relacji 4: Alicja Nowak i Weronika Jabłońska (COLLEAGUE)
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/relationship -H "Content-Type: application/json" -d '{
    "source_id": "m7",
    "target_id": "m9",
    "type": "COLLEAGUE",
    "details": "Współpraca w dziale wiadomości"
}'
check_status $?
echo "Dodano relację między m7 a m9"

# Dodawanie relacji 5: Dominik Lewandowski i Filip Górski (COLLEAGUE)
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/relationship -H "Content-Type: application/json" -d '{
    "source_id": "m8",
    "target_id": "m10",
    "type": "COLLEAGUE",
    "details": "Współpraca w komisji spraw zagranicznych"
}'
check_status $?
echo "Dodano relację między m8 a m10"

# Dodawanie relacji 6: Ewa Malinowska i Natalia Kowalczyk (FAMILY)
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/relationship -H "Content-Type: application/json" -d '{
    "source_id": "m1",
    "target_id": "m3",
    "type": "FAMILY",
    "details": "Siostra"
}'
check_status $?
echo "Dodano relację między m1 a m3"

# Dodawanie relacji 7: Jakub Szymański i Monika Zielińska (FAMILY)
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/relationship -H "Content-Type: application/json" -d '{
    "source_id": "m2",
    "target_id": "m5",
    "type": "FAMILY",
    "details": "Małżeństwo"
}'
check_status $?
echo "Dodano relację między m2 a m5"

echo "Zakończono dodawanie relacji!"