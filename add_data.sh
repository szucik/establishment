#!/bin/bash

# Skrypt do dodawania 10 nowych osób do bazy Neo4j przez API z partiami PO, Lewica, PiS, KO

# Funkcja do sprawdzania odpowiedzi curl
check_status() {
    if [ $1 -ne 0 ]; then
        echo "Błąd: Polecenie curl zwróciło kod błędu $1"
        exit 1
    fi
}

echo "Dodawanie nowych osób..."

# Dodawanie osób
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m1",
    "name": "Ewa Malinowska",
    "occupation": "Polityk",
    "party": "PO",
    "sb_status": "Brak powiązań",
    "twitter": "@EwaMalinowskaPL",
    "description": "Posłanka, zajmuje się polityką społeczną"
}'
check_status $?
echo "Dodano osobę m1"

curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m2",
    "name": "Jakub Szymański",
    "occupation": "Polityk",
    "party": "PiS",
    "sb_status": "Współpracownik SB",
    "twitter": "@JSzymanski_PiS",
    "description": "Radny wojewódzki, ekspert ds. transportu"
}'
check_status $?
echo "Dodano osobę m2"

curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m3",
    "name": "Natalia Kowalczyk",
    "occupation": "Dziennikarz",
    "sb_status": "Brak powiązań",
    "twitter": "@NatiKowalczyk",
    "description": "Reporterka śledcza w radiu"
}'
check_status $?
echo "Dodano osobę m3"

curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m4",
    "name": "Grzegorz Wiśniewski",
    "occupation": "Polityk",
    "party": "Lewica",
    "sb_status": "Funkcjonariusz PRL",
    "twitter": "@G_Wisniewski_L",
    "description": "Minister środowiska"
}'
check_status $?
echo "Dodano osobę m4"

curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m5",
    "name": "Monika Zielińska",
    "occupation": "Dziennikarz",
    "sb_status": "Brak powiązań",
    "twitter": "@MonikaZielPL",
    "description": "Redaktorka polityczna w portalu internetowym"
}'
check_status $?
echo "Dodano osobę m5"

curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m6",
    "name": "Bartosz Kaczmarek",
    "occupation": "Polityk",
    "party": "KO",
    "sb_status": "Brak powiązań",
    "twitter": "@B_Kaczmarek_KO",
    "description": "Poseł, specjalista ds. energetyki"
}'
check_status $?
echo "Dodano osobę m6"

curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m7",
    "name": "Alicja Nowak",
    "occupation": "Dziennikarz",
    "sb_status": "Brak powiązań",
    "twitter": "@AlicjaNowakTV",
    "description": "Prezenterka wiadomości"
}'
check_status $?
echo "Dodano osobę m7"

curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m8",
    "name": "Dominik Lewandowski",
    "occupation": "Polityk",
    "party": "PiS",
    "sb_status": "Współpracownik SB",
    "twitter": "@D_LewandowskiPL",
    "description": "Wiceminister spraw zagranicznych"
}'
check_status $?
echo "Dodano osobę m8"

curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m9",
    "name": "Weronika Jabłońska",
    "occupation": "Dziennikarz",
    "sb_status": "Brak powiązań",
    "twitter": "@W_Jablonska",
    "description": "Publicystka kulturalna"
}'
check_status $?
echo "Dodano osobę m9"

curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/person -H "Content-Type: application/json" -d '{
    "id": "m10",
    "name": "Filip Górski",
    "occupation": "Polityk",
    "party": "KO",
    "sb_status": "Brak powiązań",
    "twitter": "@FilipGorski_KO",
    "description": "Senator, zajmuje się edukacją wyższą"
}'
check_status $?
echo "Dodano osobę m10"

echo "Zakończono dodawanie nowych osób!"