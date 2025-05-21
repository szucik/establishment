
Generated with Ai models (Grok, ChatGpt)

docker run -p7474:7474 -p7687:7687 -v $HOME/neo4j/data:/data  -d -e  NEO4J_AUTH=neo4j/secretgraph neo4j:latest

docker exec -it 6bc8864e93d4 bin/cypher-shell -u neo4j -p secretgraph

MATCH (p:Person) RETURN p.id, p.name, p.occupation, p.twitter, p.profile_picture;