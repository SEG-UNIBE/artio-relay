# artio Relay
[![CI](https://github.com/SEG-UNIBE/artio-relay/actions/workflows/docker_build.yml/badge.svg)](https://github.com/SEG-UNIBE/artio-relay/actions/workflows/docker_build.yml)
[![CI](https://github.com/SEG-UNIBE/artio-relay/actions/workflows/docker_build_publish_development.yml/badge.svg)](https://github.com/SEG-UNIBE/artio-relay/actions/workflows/docker_build_publish_development.yml)
[![CI](https://github.com/SEG-UNIBE/artio-relay/actions/workflows/docker_build_publish_main.yml/badge.svg)](https://github.com/SEG-UNIBE/artio-relay/actions/workflows/docker_build_publish_main.yml)
[![CI](https://github.com/SEG-UNIBE/artio-relay/actions/workflows/go_test.yml/badge.svg)](https://github.com/SEG-UNIBE/artio-relay/actions/workflows/go_test.yml)

![artio-relay](./identity/logo_relay.svg)

Artio Relay is the nostr relay implementation of the Software Engineering Group of the University of Bern. 

This implementation is built upon a postgres database with a custom data store. 

# Docker
In order to run the docker compose execute the following statement from the root of this project
```docker
docker compose -f .\build\docker-compose.yml up --build -d
```