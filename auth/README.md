# Auth Daemon

## Description
Auth Daemon is a microservice which provides api key authentication.
Calls coming to prem-gateway are routed by traefik forward-auth middleware to auth daemon.
Auth daemon checks if the api key is valid and if it is, it forwards the request to the appropriate service.