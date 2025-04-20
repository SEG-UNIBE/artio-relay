# artio Relay

Artio Relay is the nostr relay implementation of the Software Engineering Group of the University of Bern. 

This implementation is built upon a postgres database with a custom data store. 

# Docker
In order to run the docker compose execute the following statement from the root of this project
```
docker compose -f .\build\docker-compose.yml up --build -d
```