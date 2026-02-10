# artio Relay
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/SEG-UNIBE/artio-relay?status.svg)](https://pkg.go.dev/github.com/SEG-UNIBE/artio-relay)
[![Go Report Card](https://goreportcard.com/badge/github.com/SEG-UNIBE/artio-relay)](https://goreportcard.com/report/github.com/SEG-UNIBE/artio-relay)
[![GitHub release](https://img.shields.io/github/tag/SEG-UNIBE/artio-relay.svg?label=release)](https://github.com/SEG-UNIBE/artio-relay/releases)
[![GitHub release date](https://img.shields.io/github/release-date/SEG-UNIBE/artio-relay.svg)](https://github.com/SEG-UNIBE/artio-relay/releases)

## Actions
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

or you can always pull our latest docker image by running:
```shell
docker pull ghcr.io/seg-unibe/artio-relay:latest
```

# Standalone
For running standalone see the ```build``` folder where you can find more build scripts. 

# Deployment

The application will then be running on port 8000 by default. 
In order to forward connections you have to set up a proper proxy on your own hosts or change the docker compose accordingly. 

The use of a HA Proxy or NGINX is highly advised for proper access management. 


# Additional Information
If you have any other questions about the specifics of the implementation and the deployment process, feel free to get in touch